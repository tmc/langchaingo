package fstring

func Format(format string, values map[string]any) (string, error) {
	p := newParser(format, values)
	if err := p.parse(); err != nil {
		return "", err
	}
	return string(p.result), nil
}
