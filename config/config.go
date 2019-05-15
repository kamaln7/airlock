package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"github.com/micro/go-config/source/flag"
)

var doRegions = []string{"ams3", "nyc3", "sgp1", "sfo2", "fra1"}

type Config struct {
	SpacesAccessKey string `json:"spacesaccesskey"`
	SpacesSecret    string `json:"spacessecret"`
	Region          string `json:"region"`
	CreateIndexes   bool   `json:"createindexes"`
	CopyToClipboard bool   `json:"copytoclipboard"`
	DryRun          bool   `json:"dryrun"`
}

func Read(path string) *Config {
	// config w/ defaults
	conf := Config{
		CreateIndexes:   true,
		CopyToClipboard: true,
		DryRun:          false,
	}

	// load user config
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
