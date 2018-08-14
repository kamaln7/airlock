package airlock

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func makeFSReader(path string) FileReader {
	return func() ([]byte, error) {
		return ioutil.ReadFile(path)
	}
}

type ErrDoesNotExist error

func (a *Airlock) ScanFiles(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrDoesNotExist(err)
		} else {
			return err
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	a.tree = make(map[string]*File)

	if info.IsDir() {
		return a.scanDirectory(absPath)
	} else {
		file := &File{
			RelPath:   info.Name(),
			Name:      info.Name(),
			Read:      makeFSReader(absPath),
			IsDir:     false,
			IsNotRoot: true,
		}
		a.tree["."] = &File{
			RelPath: ".",
			Name:    info.Name(),
			Read: func() ([]byte, error) {
				return nil, nil
			},
			IsDir:     true,
			IsNotRoot: false,
			Children:  []*File{file},
		}
		a.files = append(a.files, *file)

		return nil
	}
}

func (a *Airlock) scanDirectory(dirPath string) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// get relative path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, absPath)
		if err != nil {
			return err
		}

		// create struct
		file := &File{
			RelPath:   relPath,
			Name:      info.Name(),
			IsDir:     info.IsDir(),
			Read:      makeFSReader(absPath),
			IsNotRoot: relPath != ".",
		}

		// insert into tree
		a.tree[relPath] = file
		// update parent's children
		if parent, found := a.tree[filepath.Dir(relPath)]; found {
			parent.Children = append(parent.Children, file)
		}

		// insert only files to a.files
		if !info.IsDir() {
			a.files = append(a.files, *file)
		}

		return nil
	})

	return err
}
