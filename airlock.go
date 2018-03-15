package airlock

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goamz/goamz/s3"
)

type Airlock struct {
	Spaces *s3.S3
	Name   string

	files []File
	space *s3.Bucket
}

func New(spaces *s3.S3, path string) (*Airlock, error) {
	name := filepath.Base(path)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, ErrDoesNotExist(err)
	}

	if info.IsDir() {
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}

	airlock := &Airlock{
		Spaces: spaces,
		Name:   name,
	}

	rand.Seed(time.Now().UTC().UnixNano())

	err = airlock.ScanFiles(path)
	if err != nil {
		return nil, err
	}

	return airlock, nil
}
