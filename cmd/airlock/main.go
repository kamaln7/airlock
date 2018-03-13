package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
	"github.com/gosuri/uiprogress"
	"github.com/kamaln7/airlock"
)

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func main() {
	endpoint := "https://nyc3.digitaloceanspaces.com"
	accessKey := os.Getenv("SPACES_ACCESS_KEY")
	secretAccessKey := os.Getenv("SPACES_SECRET")

	if len(os.Args) < 2 {
		log.Fatalln("Usage: airlock <directory>")
	}

	directory := os.Args[1]
	if stat, err := os.Stat(directory); os.IsNotExist(err) || !stat.IsDir() {
		log.Fatalf("%s is not a directory", directory)
	}

	rootDirPath, err := filepath.Abs(directory)
	if err != nil {
		log.Fatalln(err)
	}
	rootDirName := filepath.Base(rootDirPath)

	fmt.Printf("\t* uploading %s to spaces\n", rootDirName)

	spaces := connectSpaces(endpoint, accessKey, secretAccessKey)

	space, err := generateSpace(rootDirName, spaces)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("\t* created Space %s\n", space.Name)

	fmt.Printf("\t* indexing files\n")
	files, err := indexFiles(directory, rootDirPath)

	fmt.Printf("\t* uploading files\n\n")
	uploadFiles(files, space)

	fmt.Printf("\n\t-> https://%s.nyc3.digitaloceanspaces.com", space.Name)
}

func indexFiles(path, rootDirPath string) ([]airlock.File, error) {
	var files []airlock.File

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		var file airlock.File

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(rootDirPath, absPath)
		if err != nil {
			return err
		}

		file.AbsPath = absPath
		file.RelPath = relPath
		file.Info = info

		files = append(files, file)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func connectSpaces(endpoint, accessKey, secretAccessKey string) *s3.S3 {
	return s3.New(aws.Auth{
		AccessKey: accessKey,
		SecretKey: secretAccessKey,
	}, aws.Region{
		S3Endpoint: endpoint,
	})
}

func generateSpace(baseName string, spaces *s3.S3) (*s3.Bucket, error) {
	for {
		name := fmt.Sprintf("%s-%s", baseName, randomString(5))
		space := spaces.Bucket(name)
		err := space.PutBucket(s3.PublicRead)
		if err != nil {
			if serr, ok := err.(*s3.Error); ok && serr.Code == "BucketAlreadyExists" {
				continue
			} else {
				return nil, err
			}
		} else {
			return space, nil
		}
	}
}

func uploadFiles(files []airlock.File, space *s3.Bucket) {
	uiprogress.Start()
	bar := uiprogress.AddBar(len(files))
	bar.AppendCompleted()
	bar.PrependElapsed()

	var err error
	for _, file := range files {
		bar.Incr()
		err = file.Upload(space)

		if err != nil {
			break
		}
	}
	if err != nil {
		log.Fatalln(err)
	}
}
