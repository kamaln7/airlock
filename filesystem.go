package airlock

import (
	"os"
	"path/filepath"
)

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
		// uploading only one file, create the file "tree" manually
		descriptor, err := os.Open(path)
		if err != nil {
			return err
		}

		file := &File{
			RelPath:   info.Name(),
			Name:      info.Name(),
			IsDir:     false,
			IsNotRoot: true,
			Size:      info.Size(),
			Reader:    descriptor,
		}
		a.tree["."] = &File{
			RelPath:   ".",
			Name:      info.Name(),
			IsDir:     true,
			IsNotRoot: false,
			Children:  []*File{file},
			Size:      info.Size(),
			Reader:    descriptor,
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

		descriptor, err := os.Open(path)
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
			IsNotRoot: relPath != ".",
			Reader:    descriptor,
			Size:      info.Size(),
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
