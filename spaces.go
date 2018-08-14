package airlock

import (
	"fmt"
	"math/rand"
	"sync"
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
	var nBars int
	if len(a.files) < 3 {
		nBars = 1
	} else {
		nBars = 3
	}

	wgs := a.makeWorkGroups(nBars)

	p := uiprogress.New()
	p.Start()

	var (
		waitGroup sync.WaitGroup
		errChan   = make(chan error, 1)
	)

	for _, wg := range wgs {
		waitGroup.Add(1)
		go wg.Work(&waitGroup, errChan, a.space, p)
	}
	waitGroup.Wait()

	close(errChan)
	p.Stop()

	return <-errChan
}
