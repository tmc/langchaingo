package tools

type Tool interface {
	Name() string
	Description() string
	Run(query string) string
}
