package airlock

import (
	"io"

	"github.com/gabriel-vasile/mimetype"
)

type File struct {
	RelPath, Name    string
	IsDir, IsNotRoot bool
	Children         []*File
	Reader           io.ReadSeeker
	Size             int64

	uploadTries int
	contentType string // if set, it is used instead of reading the file
}

func (f *File) ContentType() (string, error) {
	if f.contentType != "" {
		return f.contentType, nil
	}

	contentType, _, err := mimetype.DetectReader(f.Reader)
	f.Reader.Seek(0, io.SeekStart)

	return contentType, err
}
