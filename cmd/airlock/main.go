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

// we currently need to add index.html to URIs manually
const appendIndexURI = true

// version is filled in using ldflags
var version = "v-dev"

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

	// set usage function
	flag.Usage = func() {
		fmt.Printf(`%s %s

Usage: airlock <path>

Arguments:
	-config string
		path to airlock config file (default "%s")
	-region string
		Spaces region
	-spacesAccessKey string
		Spaces access key
	-spacesSecret string
		Spaces secret
	-copyToClipboard
		Copy the resulting Spaces URL to the clipboard (default true)
	-createIndexes
		Create index.html files or not (default true)
	-dryRun
		Test run without contacting Spaces at all
`, color.New(color.FgBlue).Sprint("airlock"), version, homedirPath)
	}

	// read config path from flag
	flag.StringVar(&configPath, "config", homedirPath, "path to airlock config file")
	flag.String("spacesAccessKey", "", "Spaces access key")
	flag.String("spacesSecret", "", "Spaces secret")
	flag.String("region", "", "Spaces region")
	flag.Bool("createIndexes", true, "Create index.html files or not")
	flag.Bool("copyToClipboard", true, "Copy the resulting Spaces URL to the clipboard")
	flag.Bool("dryRun", false, "Test run without contacting Spaces at all")
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

	arg := flag.Arg(0)
	if arg == "" {
		flag.Usage()

		os.Exit(0)
	}

	fmt.Printf("\t🌌 connecting to Spaces\n")
	endpoint := fmt.Sprintf("https://%s.digitaloceanspaces.com", conf.Region)
	spaces := connectSpaces(endpoint, conf.SpacesAccessKey, conf.SpacesSecret)

	fmt.Println("\t⚙  indexing files")
	al, err := airlock.New(spaces, arg)
	if err != nil {
		log.Fatalln(err)
	}

	al.DryRun = conf.DryRun

	if conf.CreateIndexes {
		err = al.AddFileListings()
		if err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Println("\t⚙  creating Space")
	err = al.MakeSpace()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("\t➕ created Space %s\n", color.BlueString(al.SpaceName()))

	fmt.Printf("\t✏️  uploading files\n\n")
	err = al.Upload()
	if err != nil {
		if serr, ok := err.(*s3.Error); ok {
			fmt.Printf("%#v\n", serr)
		} else {
			log.Fatalln(err)
		}
	}

	url := fmt.Sprintf("https://%s.%s.digitaloceanspaces.com", al.SpaceName(), conf.Region)
	if appendIndexURI {
		url = url + "/index.html"
	}

	fmt.Printf("\n\t🚀 %s\n", color.New(color.FgBlue).Sprint(url))

	if conf.CopyToClipboard {
		clipboard.WriteAll(url)
		fmt.Printf("\t  %s\n", color.New(color.FgHiBlack).Sprint(" (in clipboard)"))
	}
}
