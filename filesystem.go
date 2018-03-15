package airlock

import (
	"os"
	"path/filepath"
)

type ErrDoesNotExist error

func (a *Airlock) ScanFiles(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ErrDoesNotExist(err)
	}

	if info.IsDir() {
		// recursive scan
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		return a.scanDirectory(absPath)
	} else {
		// one file
		file := File{
			Info:    info,
			Path:    path,
			RelPath: filepath.Base(path),
		}

		a.files = append(a.files, file)
		return nil
	}
}

func (a *Airlock) scanDirectory(dirPath string) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, absPath)
		if err != nil {
			return err
		}

		file := File{
			Path:    path,
			RelPath: relPath,
			Info:    info,
		}

		a.files = append(a.files, file)

		return nil
	})

	return err
}
