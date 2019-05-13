package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"github.com/micro/go-config/source/flag"
)

var doRegions = []string{"ams3", "nyc3", "sgp1", "sfo2", "fra1"}

type Config struct {
	SpacesAccessKey string `json:"spacesAccessKey"`
	SpacesSecret    string `json:"spacesSecret"`
	Region          string `json:"region"`
	CreateIndexes   bool   `json:"createIndexes"`
	CopyToClipboard bool   `json:"copyToClipboard"`
	DryRun          bool   `json:"dryRun"`
}

func Read(path string) *Config {
	config.Load(
		// base from file
		file.NewSource(
			file.WithPath(path),
		),
		// override with env
		env.NewSource(),
		// override with flags
		flag.NewSource(),
	)

	var conf Config

	// set defaults
	conf.CreateIndexes = true
	conf.CopyToClipboard = true
	conf.DryRun = false

	// scan config into struct
	config.Scan(&conf)

	return &conf
}

func (c *Config) Validate() bool {
	valid := true

	valid = valid && c.SpacesAccessKey != ""
	valid = valid && c.SpacesSecret != ""
	valid = valid && validRegion(c.Region)

	return valid
}

func validRegion(region string) bool {
	for _, v := range doRegions {
		if region == v {
			return true
		}
	}

	return false
}
