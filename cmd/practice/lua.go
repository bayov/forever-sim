//go:build practice

package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// A small reader for the SavedVariables files WoW writes: plain Lua tables of strings, numbers,
// booleans and nested tables. It reads the 850 KB BayovCore file in well under a second.
//
// A table becomes a map[string]any. Array entries ({ a, b }) and numeric keys ([3] = x) get
// their number as the key, so both read the same.

type luaParser struct {
	s   string
	pos int
	// next array index of each open table
}

// parseSavedVariables reads `Name = value` assignments at the top level of a SavedVariables
// file.
func parseSavedVariables(text string) (map[string]any, error) {
	p := &luaParser{s: text}
	out := map[string]any{}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			return out, nil
		}
		name := p.ident()
		if name == "" {
			return nil, p.errorf("expected a variable name")
		}
		p.skipSpace()
		if !p.consume('=') {
			return nil, p.errorf("expected '=' after %s", name)
		}
		value, err := p.value()
		if err != nil {
			return nil, err
		}
		out[name] = value
	}
}

func (p *luaParser) errorf(format string, args ...any) error {
	line := strings.Count(p.s[:min(p.pos, len(p.s))], "\n") + 1
	return fmt.Errorf("lua line %d: %s", line, fmt.Sprintf(format, args...))
}

func (p *luaParser) skipSpace() {
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			p.pos++
		case strings.HasPrefix(p.s[p.pos:], "--"):
			if end := strings.IndexByte(p.s[p.pos:], '\n'); end >= 0 {
				p.pos += end + 1
			} else {
				p.pos = len(p.s)
			}
		default:
			return
		}
	}
}

func (p *luaParser) consume(c byte) bool {
	p.skipSpace()
	if p.pos < len(p.s) && p.s[p.pos] == c {
		p.pos++
		return true
	}
	return false
}

func (p *luaParser) ident() string {
	start := p.pos
	for p.pos < len(p.s) {
		r := rune(p.s[p.pos])
		if r == '_' || unicode.IsLetter(r) || (p.pos > start && unicode.IsDigit(r)) {
			p.pos++
			continue
		}
		break
	}
	return p.s[start:p.pos]
}

func (p *luaParser) value() (any, error) {
	p.skipSpace()
	if p.pos >= len(p.s) {
		return nil, p.errorf("unexpected end")
	}
	switch c := p.s[p.pos]; {
	case c == '{':
		return p.table()
	case c == '"' || c == '\'':
		return p.str()
	case c == '-' || c == '.' || (c >= '0' && c <= '9'):
		return p.number()
	default:
		word := p.ident()
		switch word {
		case "true":
			return true, nil
		case "false":
			return false, nil
		case "nil":
			return nil, nil
		}
		return nil, p.errorf("unexpected %q", word)
	}
}

func (p *luaParser) number() (any, error) {
	start := p.pos
	for p.pos < len(p.s) && strings.ContainsRune("-+.0123456789eExXabcdefABCDEF", rune(p.s[p.pos])) {
		p.pos++
	}
	text := p.s[start:p.pos]
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "-0x") {
		v, err := strconv.ParseInt(text, 0, 64)
		return float64(v), err
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, p.errorf("bad number %q", text)
	}
	return v, nil
}

func (p *luaParser) str() (string, error) {
	quote := p.s[p.pos]
	p.pos++
	var b strings.Builder
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		p.pos++
		switch c {
		case quote:
			return b.String(), nil
		case '\\':
			if p.pos >= len(p.s) {
				return "", p.errorf("unterminated string")
			}
			e := p.s[p.pos]
			p.pos++
			switch e {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '\n':
				b.WriteByte('\n')
			default:
				if e >= '0' && e <= '9' {
					// \ddd, a decimal byte
					end := p.pos - 1
					for end < len(p.s) && end < p.pos+2 && p.s[end] >= '0' && p.s[end] <= '9' {
						end++
					}
					n, _ := strconv.Atoi(p.s[p.pos-1 : end])
					b.WriteByte(byte(n))
					p.pos = end
				} else {
					b.WriteByte(e)
				}
			}
		default:
			b.WriteByte(c)
		}
	}
	return "", p.errorf("unterminated string")
}

func (p *luaParser) table() (map[string]any, error) {
	p.pos++ // {
	out := map[string]any{}
	next := 1
	for {
		p.skipSpace()
		if p.consume('}') {
			return out, nil
		}
		var key string
		if p.s[p.pos] == '[' {
			p.pos++
			k, err := p.value()
			if err != nil {
				return nil, err
			}
			if !p.consume(']') {
				return nil, p.errorf("expected ']'")
			}
			if !p.consume('=') {
				return nil, p.errorf("expected '='")
			}
			key = keyString(k)
		} else if r := rune(p.s[p.pos]); r == '_' || unicode.IsLetter(r) {
			save := p.pos
			word := p.ident()
			p.skipSpace()
			if word != "true" && word != "false" && word != "nil" && p.consume('=') {
				key = word
			} else {
				p.pos = save
			}
		}
		value, err := p.value()
		if err != nil {
			return nil, err
		}
		if key == "" {
			key = strconv.Itoa(next)
			next++
		}
		out[key] = value
		p.skipSpace()
		if !p.consume(',') && !p.consume(';') {
			p.skipSpace()
			if p.consume('}') {
				return out, nil
			}
			return nil, p.errorf("expected ',' or '}'")
		}
	}
}

func keyString(k any) string {
	switch v := k.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	}
	return fmt.Sprint(k)
}

// Helpers to walk the parsed tables.

func luaTable(v any, path ...string) map[string]any {
	for _, key := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[key]
	}
	m, _ := v.(map[string]any)
	return m
}

func luaString(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

func luaNumber(m map[string]any, key string) float64 {
	n, _ := m[key].(float64)
	return n
}

func luaBool(m map[string]any, key string) bool {
	b, _ := m[key].(bool)
	return b
}
