package airlock

import (
	"fmt"
	"math/rand"
	"time"

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
