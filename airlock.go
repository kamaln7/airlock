package airlock

import (
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goamz/goamz/s3"
)

type Airlock struct {
	Spaces *s3.S3

	name        string
	files       []File
	tree        map[string]*File
	space       *s3.Bucket
	listingTmpl *template.Template
}

var (
	SpaceNameRegexp       = regexp.MustCompile(`[^a-z0-9\-]+`)
	SpaceNamePrefixRegexp = regexp.MustCompile(`[^a-z0-9]`)
)

const (
	SpaceNameMaxLength  = 63
	SpaceNameRandLength = 5
)

func New(spaces *s3.S3, path string) (*Airlock, error) {
	al := &Airlock{
		Spaces: spaces,
	}

	err := al.SetName(path)

	err = al.ScanFiles(path)
	if err != nil {
		return nil, err
	}

	return al, nil
}

func (a *Airlock) SetName(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrDoesNotExist(err)
		}

		return err
	}

	// use absolute path to include the directory's name in case for example "." is passed as the path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	name := filepath.Base(absPath)

	if !info.IsDir() {
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}

	name = strings.ToLower(name)
	name = SpaceNameRegexp.ReplaceAllString(name, "")
	name = strings.TrimLeftFunc(name, func(r rune) bool {
		return SpaceNamePrefixRegexp.MatchString(string(r))
	})

	if len(name) == 0 {
		name = "airlock"
	}

	a.name = name
	return nil
}
