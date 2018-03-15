package airlock

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/goamz/goamz/s3"
)

type File struct {
	Path, RelPath string
	Info          os.FileInfo
}

func (f File) Upload(space *s3.Bucket) error {
	if f.Info.IsDir() {
		return nil
	}

	fileContent, err := ioutil.ReadFile(f.Path)
	if err != nil {
		return err
	}

	contentType := http.DetectContentType(fileContent)
	return space.Put(f.RelPath, fileContent, contentType, s3.PublicRead, s3.Options{})
}
