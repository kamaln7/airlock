package main

import (
	"fmt"
	"log"
	"os"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
	"github.com/kamaln7/airlock"
)

func connectSpaces(endpoint, accessKey, secretAccessKey string) *s3.S3 {
	return s3.New(aws.Auth{
		AccessKey: accessKey,
		SecretKey: secretAccessKey,
	}, aws.Region{
		S3Endpoint: endpoint,
	})
}

func main() {
	endpoint := "https://nyc3.digitaloceanspaces.com"
	accessKey := os.Getenv("SPACES_ACCESS_KEY")
	secretAccessKey := os.Getenv("SPACES_SECRET")

	if len(os.Args) < 2 {
		log.Fatalln("Usage: airlock <path>")
	}

	fmt.Printf("\t* connecting to Spaces\n")
	spaces := connectSpaces(endpoint, accessKey, secretAccessKey)

	fmt.Printf("\t* indexing files\n")
	path := os.Args[1]
	al, err := airlock.New(spaces, path)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("\t* creating Space\n")
	err = al.MakeSpace()
	if err != nil {
		log.Fatalln(err)
	}

	al.Upload()

	fmt.Printf("\n\t-> https://%s.nyc3.digitaloceanspaces.com", al.SpaceName())
}
