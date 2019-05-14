package airlock

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/goamz/goamz/s3"
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

		if a.DryRun {
			a.space = &s3.Bucket{
				Name: spaceName,
			}
			return nil
		}

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

func (a *Airlock) uploadFile(f File) error {
	// do nothing on dry run
	if a.DryRun {
		time.Sleep(time.Millisecond * 300)
		return nil
	}

	if f.IsDir {
		return nil
	}

	contentType, err := f.ContentType()
	if err != nil {
		return err
	}

	return a.space.PutReader(f.RelPath, f.Reader, f.Size, contentType, s3.PublicRead, s3.Options{})
}
