package airlock

import (
	"net/http"

	"github.com/goamz/goamz/s3"
)

type FileReader func() ([]byte, error)

type File struct {
	RelPath, Name    string
	IsDir, IsNotRoot bool
	Read             FileReader
	Children         []*File
}

func (f *File) Upload(space *s3.Bucket) error {
	if f.IsDir {
		return nil
	}

	content, err := f.Read()
	if err != nil {
		return err
	}

	contentType := http.DetectContentType(content)
	return space.Put(f.RelPath, content, contentType, s3.PublicRead, s3.Options{})
}
