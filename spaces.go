package airlock

import (
	"fmt"
	"math/rand"

	"github.com/goamz/goamz/s3"
	"github.com/gosuri/uiprogress"
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
	for {
		name := fmt.Sprintf("%s-%s", a.Name, randomString(5))
		space := a.Spaces.Bucket(name)
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

	var err error
	for _, file := range a.files {
		err = file.Upload(a.space)

		bar.Incr()
		if err != nil {
			break
		}
	}

	return err
}
