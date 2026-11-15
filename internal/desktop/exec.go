package desktop

import (
	"fmt"
	"strings"
)

// Field codes defined by the Desktop Entry Specification.
const (
	codeFile     = 'f' // single file
	codeFileList = 'F' // list of files
	codeURL      = 'u' // single URL
	codeURLList  = 'U' // list of URLs
	codeIcon     = 'i' // --icon and the icon value
	codeName     = 'c' // translated name
	codeLocation = 'k' // file location or URI
	codePercent  = '%' // literal percent sign
)

// deprecatedCodes are the field codes the specification marks as
// deprecated. They behave like their modern counterparts.
var deprecatedCodes = map[byte]byte{
	'd': codeFile,
	'D': codeFileList,
	'n': codeIcon,
	'N': codeIcon,
	'v': codeName,
	'm': codeIcon,
}

// ExpandOptions carries the values a command line is expanded with.
type ExpandOptions struct {
	// Name is the entry's localised name, used by %c.
	Name string
	// Icon is the icon path, used by %i.
	Icon string
	// Files are the file arguments, used by %f and %F.
	Files []string
	// URLs are the URL arguments, used by %u and %U.
	URLs []string
	// Location is the file location, used by %k.
	Location string
}

// Expand turns an Exec value into an argument list.
//
// Field codes that cannot be satisfied are removed, together with the
// argument that holds them, as the specification requires.
func Expand(exec string, opts ExpandOptions) ([]string, error) {
	tokens, err := tokenize(exec)
	if err != nil {
		return nil, err
	}

	var args []string
	for _, token := range tokens {
		expanded, keep := expandToken(token, opts)
		if !keep {
			continue
		}
		args = append(args, expanded...)
	}
	return args, nil
}

// codePos records where a field code appears in a raw argument.
type codePos struct {
	start int
	end   int
	code  byte
}

// token is a single argument of an Exec value.
type token struct {
	// raw is the argument as written, including quotes and escapes.
	raw string
	// codes holds the field codes found in the argument, in order.
	codes []codePos
}

// isCodeOnly reports whether the argument consists of a single field code.
func (t token) isCodeOnly() bool {
	if len(t.codes) != 1 {
		return false
	}
	return stripQuotes(t.raw) == fmt.Sprintf("%%%c", t.codes[0].code)
}

// tokenize splits an Exec value into arguments. Quoted sections keep their
// spaces and lose their quotes, and a backslash escapes the next character.
func tokenize(exec string) ([]token, error) {
	var tokens []token
	var current strings.Builder
	var codes []codePos
	started := false
	inQuotes := false
	escaped := false

	flush := func() {
		if !started {
			return
		}
		tokens = append(tokens, token{raw: current.String(), codes: codes})
		current.Reset()
		codes = nil
		started = false
	}

	for i := 0; i < len(exec); i++ {
		c := exec[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		switch c {
		case '\\':
			if !inQuotes {
				// Keep the escape so it can be resolved later.
				current.WriteByte(c)
			}
			escaped = true
			started = true
			continue
		case '"':
			inQuotes = !inQuotes
			started = true
			continue
		case ' ':
			if inQuotes {
				current.WriteByte(c)
				continue
			}
			flush()
			continue
		case '%':
			if i+1 >= len(exec) {
				return nil, fmt.Errorf("desktop: trailing %% in Exec value")
			}
			code := exec[i+1]
			if code == codePercent {
				current.WriteByte('%')
				started = true
				i++
				continue
			}
			if replacement, ok := deprecatedCodes[code]; ok {
				code = replacement
			}
			codes = append(codes, codePos{start: current.Len(), end: current.Len() + 2, code: code})
			current.WriteString("%" + string(code))
			i++
			started = true
			continue
		}

		current.WriteByte(c)
		started = true
	}

	if escaped {
		return nil, fmt.Errorf("desktop: trailing escape in Exec value")
	}
	if inQuotes {
		return nil, fmt.Errorf("desktop: unbalanced quote in Exec value")
	}
	flush()

	return tokens, nil
}

// expandToken expands a single argument.
func expandToken(t token, opts ExpandOptions) ([]string, bool) {
	if len(t.codes) == 0 {
		return []string{resolve(t.raw)}, true
	}

	if t.isCodeOnly() {
		switch t.codes[0].code {
		case codeFileList:
			return opts.Files, len(opts.Files) > 0
		case codeURLList:
			return opts.URLs, len(opts.URLs) > 0
		case codeFile:
			return opts.Files[:min(1, len(opts.Files))], len(opts.Files) > 0
		case codeURL:
			return opts.URLs[:min(1, len(opts.URLs))], len(opts.URLs) > 0
		case codeIcon:
			if opts.Icon == "" {
				return nil, false
			}
			return []string{"--icon", opts.Icon}, true
		case codeName:
			if opts.Name == "" {
				return nil, false
			}
			return []string{opts.Name}, true
		case codeLocation:
			if opts.Location == "" {
				return nil, false
			}
			return []string{opts.Location}, true
		}
	}

	value, ok := buildArg(t, opts)
	if !ok || value == "" {
		return nil, false
	}
	return []string{value}, true
}

// buildArg resolves quotes and escapes and substitutes the field codes of
// an argument that mixes text with codes.
func buildArg(t token, opts ExpandOptions) (string, bool) {
	var b strings.Builder
	inQuotes := false
	escaped := false
	next := 0

	for i := 0; i < len(t.raw); i++ {
		if next < len(t.codes) && i == t.codes[next].start {
			value, ok := codeValue(t.codes[next].code, opts)
			if !ok {
				return "", false
			}
			b.WriteString(value)
			i = t.codes[next].end - 1
			next++
			continue
		}

		c := t.raw[i]
		switch {
		case escaped:
			b.WriteByte(c)
			escaped = false
		case c == '\\' && !inQuotes:
			escaped = true
		case c == '\\':
			b.WriteByte(c)
		case c == '"':
			inQuotes = !inQuotes
		default:
			b.WriteByte(c)
		}
	}
	return b.String(), true
}

// codeValue returns the value a field code expands to. It reports false
// when the code has no value, which removes the argument.
func codeValue(code byte, opts ExpandOptions) (string, bool) {
	switch code {
	case codeFileList, codeFile:
		return first(opts.Files), true
	case codeURLList, codeURL:
		return first(opts.URLs), true
	case codeIcon:
		return opts.Icon, true
	case codeName:
		return opts.Name, true
	case codeLocation:
		return opts.Location, true
	}
	return "", false
}

// resolve removes quotes and resolves escapes of an argument without field
// codes.
func resolve(raw string) string {
	value, _ := buildArg(token{raw: raw}, ExpandOptions{})
	return value
}

// stripQuotes removes quotes and whitespace from an argument.
func stripQuotes(raw string) string {
	var b strings.Builder
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case '"', ' ', '\t':
		case '\\':
			if i+1 < len(raw) {
				i++
				b.WriteByte(raw[i])
			}
		default:
			b.WriteByte(raw[i])
		}
	}
	return b.String()
}

// first returns the first element of a list or an empty string.
func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
