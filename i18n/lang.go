package i18n

// Lang is the language type.
type Lang int

const (
	// EN stands for English.
	EN Lang = iota
	// ZH stands for Simplified-Chinese.
	ZH

	// DefaultLang is the default language.
	DefaultLang = EN
)
