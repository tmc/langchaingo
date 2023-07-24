package load

type Serializer interface {
	ToFile(data any, path string) error
	FromFile(data any, path string) error
}
