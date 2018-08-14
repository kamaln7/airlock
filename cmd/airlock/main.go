package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
	"github.com/mitchellh/go-homedir"

	"github.com/kamaln7/airlock"
	"github.com/kamaln7/airlock/config"
)

// version is filled in using ldflags
var version = "-dev"

func connectSpaces(endpoint, accessKey, secretAccessKey string) *s3.S3 {
	return s3.New(aws.Auth{
		AccessKey: accessKey,
		SecretKey: secretAccessKey,
	}, aws.Region{
		S3Endpoint: endpoint,
	})
}

var configPath string

func setOptions() {
	homedirPath, err := homedir.Dir()
	if err != nil {
		log.Printf("could not find home directory: %v\n", err)
		homedirPath = ""
	} else {
		homedirPath = filepath.Join(homedirPath, ".airlock.yaml")
	}

	// read config path from flag
	flag.StringVar(&configPath, "config", homedirPath, "path to airlock config file")
	flag.Parse()

	// override with env
	envPath := os.Getenv("CONFIG")
	if envPath != "" {
		configPath = envPath
	}
}

func main() {
	setOptions()

	conf := config.Read(configPath)
	if !conf.Validate() {
		log.Fatalln("config is invalid.")
	}

	if len(os.Args) < 2 || os.Args[1] == "version" || os.Args[1] == "help" {
		fmt.Printf(`%s %s
	Usage: airlock <path>
`, color.New(color.FgBlue).Sprint("airlock"), version)

		os.Exit(0)
	}

	fmt.Printf("\t🌌 connecting to Spaces\n")
	endpoint := fmt.Sprintf("https://%s.digitaloceanspaces.com", conf.Region)
	spaces := connectSpaces(endpoint, conf.SpacesAccessKey, conf.SpacesSecret)

	fmt.Println("\t🌌 indexing files")
	path := os.Args[1]
	al, err := airlock.New(spaces, path)
	if err != nil {
		log.Fatalln(err)
	}

	if conf.CreateIndexes {
		err = al.AddFileListings()
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Println("\t🌌 creating Space")
	err = al.MakeSpace()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("\t🌌 created Space %s\n", color.BlueString(al.SpaceName()))

	fmt.Printf("\t🌌 uploading files\n\n")
	err = al.Upload()
	if err != nil {
		if serr, ok := err.(*s3.Error); ok {
			fmt.Printf("%#v\n", serr)
		}
		log.Fatalln(err)
	}

	url := fmt.Sprintf("https://%s.%s.digitaloceanspaces.com", al.SpaceName(), conf.Region)

	fmt.Printf("\n\t🚀 %s\n", color.New(color.FgBlue).Sprint(url))

	if conf.CopyToClipboard {
		clipboard.WriteAll(url)
		fmt.Printf("\t  %s\n", color.New(color.FgHiBlack).Sprint(" (in clipboard)"))
	}
}
