package airlock

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
	"github.com/goamz/goamz/s3"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (a *Airlock) MakeSpace() error {
	rand.Seed(time.Now().UTC().UnixNano())

	cleanName := a.name
	preRandLength := SpaceNameMaxLength - SpaceNameRandLength - 1
	if len(cleanName) > preRandLength {
		cleanName = cleanName[:preRandLength]
	}

	for {
		spaceName := fmt.Sprintf("%s-%s", cleanName, randomString(SpaceNameRandLength))

		space := a.Spaces.Bucket(spaceName)
		err := space.PutBucket(s3.Private)
		if err != nil {
			if serr, ok := err.(*s3.Error); ok && serr.Code == "BucketAlreadyExists" {
				continue
			} else {
				return err
			}
		} else {
			a.space = space
			return nil
		}
	}
}

func (a *Airlock) SpaceName() string {
	return a.space.Name
}

func (a *Airlock) Upload() error {
	uiprogress.Start()
	bar := uiprogress.AddBar(len(a.files))
	bar.AppendCompleted()
	bar.PrependElapsed()
	bar.Width = 25

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		index := max(0, min(len(a.files)-1, b.Current()))

		return color.New(color.FgBlue).Sprint(strutil.Resize(" "+a.files[index].Name, 15))
	})

	var (
		err error
		// use our own counter instead of bar.Incr()
		// it does not need to be thread safe. keep track
		// of the progress so we can manually set it later
		// and force the bar to update
		progress = 1
	)
	for _, file := range a.files {
		err = file.Upload(a.space)

		bar.Set(progress)
		if err != nil {
			break
		}

		progress++
	}
	bar.Set(progress)
	uiprogress.Stop()

	return err
}
