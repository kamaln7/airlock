package airlock

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/goamz/goamz/s3"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/pkg/errors"
)

type Airlock struct {
	Spaces *s3.S3
	DryRun bool

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
	// FileUploadMaxTries is the maximum amount of times airlock will try to upload a file and receive an error before giving up on it
	FileUploadMaxTries = 2
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

func (a *Airlock) Upload() error {
	var numWorkers int
	if len(a.files) < 3 {
		numWorkers = 1
	} else {
		numWorkers = 3
	}

	// run workers and wait for them to finish
	var (
		workerWg, fileWg sync.WaitGroup
		errChan          = make(chan error)
		fileChan         = make(chan File, numWorkers)
	)

	// copy files to a files channel
	go func() {
		fileWg.Add(len(a.files))
		for _, file := range a.files {
			fileChan <- file
		}

		// close the channel only when all files have been fully processed
		fileWg.Wait()
		close(fileChan)
	}()

	// start progress bar instance
	p := uiprogress.New()
	p.Start()

	// print any received errors
	go func() {
		for err := range errChan {
			fmt.Fprintln(p.Bypass(), err.Error())
		}
	}()

	// create ui progress bars for workers and run them
	workerWg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		bar := p.AddBar(len(a.files))
		bar.Width = 30
		bar.AppendCompleted()
		go a.uploadWorker(&workerWg, &fileWg, fileChan, errChan, bar)
	}
	workerWg.Wait()

	close(errChan)
	p.Stop()

	return nil
}

func (a *Airlock) uploadWorker(workerWg, fileWg *sync.WaitGroup, fileChan chan File, errChan chan<- error, bar *uiprogress.Bar) {
	var currentFileName string // we need to keep track of this outside the loop so we can print it with the bar

	defer func() {
		currentFileName = ""
		for bar.Incr() {
		}
		workerWg.Done()
	}()

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		var (
			elapsedTime = ""
			c           *color.Color
		)

		// check if completed
		if b.Current() == b.Total {
			c = color.New(color.FgHiBlack)
			elapsedTime = "done"
		} else {
			c = color.New(color.FgBlue)
			elapsedTime = b.TimeElapsedString()
		}

		// trim strings to not take up too much cli space
		cfn := c.Sprint(strutil.Resize(currentFileName, 15))
		elapsedTime = color.New(color.FgYellow).Sprint(strutil.PadLeft(elapsedTime, 5, ' '))

		return cfn + elapsedTime
	})

	// do the magic
	for file := range fileChan {
		file.uploadTries++

		currentFileName = file.Name
		err := a.uploadFile(file)
		currentFileName = ""

		if err == nil {
			// thank u, next
			bar.Incr()
			fileWg.Done()

			if file.uploadTries != 1 {
				// this isn't the first try
				errChan <- fmt.Errorf("successfully uploaded %s after %d attempts", file.RelPath, file.uploadTries)
			}

			continue
		}

		errChan <- errors.Wrapf(err, "failed to upload %s", file.RelPath)
		// re-insert into the channel if the upload failed or ignore if hit max number of tries
		if file.uploadTries < FileUploadMaxTries {
			fileChan <- file
		} else {
			fileWg.Done()
			errChan <- fmt.Errorf("failed to upload %s after %d tries", file.RelPath, file.uploadTries)
		}
	}
}
