package airlock

type FileReader func() ([]byte, error)

type File struct {
	RelPath, Name    string
	IsDir, IsNotRoot bool
	Read             FileReader
	Children         []*File

	uploadTries int
}
