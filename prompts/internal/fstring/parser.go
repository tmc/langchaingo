package fstring

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	data   []rune
	result []rune
	idx    int
	values map[string]any
}

func newParser(s string, values map[string]any) *parser {
	if len(values) == 0 {
		values = map[string]any{}
	}
	return &parser{
		data:   []rune(s),
		result: nil,
		idx:    0,
		values: values,
	}
}

func (r *parser) parse() error {
	for r.hasMore() {
		existLeftCurlyBracket, tmp, err := r.scanToLeftCurlyBracket()
		if err != nil {
			return err
		}
		r.result = append(r.result, tmp...)
		if !existLeftCurlyBracket {
			continue
		}

		tmp, err = r.scanToRightCurlyBracket()
		if err != nil {
			return err
		}
		valName := strings.TrimSpace(string(tmp))
		if valName == "" {
			return fmt.Errorf("empty expression not allowed")
		}
		val, ok := r.values[valName]
		if !ok {
			return fmt.Errorf("name '%s' is not defined", valName)
		}
		r.result = append(r.result, []rune(toString(val))...)
	}
	return nil
}

func (r *parser) scanToLeftCurlyBracket() (bool, []rune, error) {
	res := []rune{}
	exist := false
	for r.hasMore() {
		s := r.get()
		r.idx++
		if s == '}' {
			if r.hasMore() && r.get() == '}' {
				res = append(res, '}')
				r.idx++
				continue
			}
			return false, nil, fmt.Errorf("single '}' is not allowed")
		} else if s != '{' {
			res = append(res, s)
			continue
		} else {
			if !r.hasMore() {
				return false, nil, fmt.Errorf("single '{' is not allowed")
			}
			if r.get() == '{' {
				// {{ -> {
				r.idx++
				res = append(res, '{')
				continue
			}
			exist = true
			break
		}
	}
	return exist, res, nil
}

func (r *parser) scanToRightCurlyBracket() ([]rune, error) {
	var res []rune
	for r.hasMore() {
		s := r.get()
		if s != '}' {
			// xxx
			res = append(res, s)
			r.idx++
			continue
		}
		r.idx++
		break
	}
	return res, nil
}

func (r *parser) hasMore() bool {
	return r.idx < len(r.data)
}

func (r *parser) get() rune {
	return r.data[r.idx]
}

func toString(val any) string {
	if val == nil {
		return "nil" // f'None' -> "None"
	}
	switch val := val.(type) {
	case string:
		return val
	case []rune:
		return string(val)
	case []byte:
		return string(val)
	case int:
		return strconv.FormatInt(int64(val), 10)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(int64(val), 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(uint64(val), 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(float64(val), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%s", val)
	}
}
