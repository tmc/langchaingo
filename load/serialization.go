package load

type Serializer interface {
	ToFile(data any, path string) error
	FromFile(data any, path string) error
}

type FileSerializer struct {
	FileSystem FileSystem
}

func NewSerializer(fs FileSystem) *FileSerializer {
	return &FileSerializer{
		FileSystem: fs,
	}
}
