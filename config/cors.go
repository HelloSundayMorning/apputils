package config

import (
	"fmt"
	"github.com/HelloSundayMorning/apputils/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const (
	defaultCorsUrlConfigPath = "/app/config/cors_url_config.yaml"
	corsUrlPath = "CORS_CONFIG_FILE"
)

type (
	CorsUrl struct {
		URL string `yaml:"url"`
	}

	// Cors
	// YAML config for CORS URLs. Example
	//
	//	urls:
	//		- url: "http://www.example.com"
	//		- url: "http://www.example2.com"
	//
	//
	Cors struct {
		Urls []CorsUrl `yaml:"urls"`
	}
)

// SetCorsConfig
// Will load CORS configuration from the default YAML file in /app/config/cors_url_config.yaml
// or the path in CORS_CONFIG_FILE if present.
func (config Configuration) LoadCorsConfig() (err error) {
	component := "LoadCorsConfig"

	filePath := os.Getenv(corsUrlPath)

	if filePath == "" {
		log.PrintfNoContext(config.AppID, component, "Cannot find env variable %s. Loading config from default path %s", corsUrlPath, defaultCorsUrlConfigPath)
		filePath = defaultCorsUrlConfigPath
	}

	yamlFromFile, err := ioutil.ReadFile(filePath)

	if err != nil {
		return fmt.Errorf("cannot read file %s, %s", filePath, err)
	}

	log.PrintfNoContext(config.AppID, component, "Reading YAML config file %s", filePath)

	var cors Cors

	err = yaml.Unmarshal(yamlFromFile, &cors)

	if err != nil {
		return fmt.Errorf("error reading YAML config from %s, %s", filePath, err)
	}

	config.CorsConfig = cors

	log.PrintfNoContext(config.AppID, component, "Loaded YAML config file %s", filePath)

	return nil
}

// ReadAsString
// Returns the Cors Urls in a string slice for convenience
func (cors Cors) ReadAsString() (urls []string) {

	for _, u := range cors.Urls {
		urls = append(urls, u.URL)
	}

	return urls
}
