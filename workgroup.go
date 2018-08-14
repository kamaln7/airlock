package airlock

import (
	"sync"

	"github.com/fatih/color"
	"github.com/goamz/goamz/s3"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

type workGroup struct {
	Bar   *uiprogress.Bar
	Files []File
}

func (wg *workGroup) Work(waitGroup *sync.WaitGroup, errChan chan error, space *s3.Bucket, p *uiprogress.Progress) {
	defer waitGroup.Done()

	wg.Bar = p.AddBar(len(wg.Files))
	wg.Bar.AppendCompleted()
	wg.Bar.Width = 25

	// elapsed time
	// current file name
	wg.Bar.PrependFunc(func(b *uiprogress.Bar) string {

		var (
			currentFileName = ""
			elapsedTime     = ""
			c               *color.Color
		)
		// check if completed
		if b.Current() == b.Total {
			// current file name
			currentFileName = "        done"
			c = color.New(color.FgHiBlack)

			// elapsed time
			elapsedTime = ""
		} else {
			// current file name
			index := max(0, min(len(wg.Files)-1, b.Current()))
			currentFileName = " " + wg.Files[index].Name
			c = color.New(color.FgBlue)

			// elapsed time
			elapsedTime = b.TimeElapsedString()
		}

		currentFileName = c.Sprint(strutil.Resize(currentFileName, 15))
		elapsedTime = color.New(color.FgYellow).Sprint(strutil.PadLeft(elapsedTime, 5, ' '))

		return currentFileName + elapsedTime
	})

	for _, file := range wg.Files {
		err := file.Upload(space)

		wg.Bar.Incr()
		if err != nil {
			errChan <- err
			break
		}
	}

	for wg.Bar.Incr() {
	}
}

func (a *Airlock) makeWorkGroups(n int) []*workGroup {
	var wgs []*workGroup

	// split files into work groups
	chunkSize := len(a.files) / n
	for i, k := 0, 0; i < n; i, k = i+1, k+chunkSize {
		end := k + chunkSize
		if i == n-1 {
			end = len(a.files)
		}

		// add files
		wgs = append(wgs, &workGroup{
			Files: a.files[k:end],
		})
	}

	return wgs
}
