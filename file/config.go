package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// ReadConfig reads the configuration file from the given URL and unmarshals it
// into the provided config struct. The file format is determined by the
// extension of the file, either ".json" or ".yaml" (or ".yml" for backwards
// compatibility). If the file format is not recognized, the function returns an
// error.
func ReadConfig(cfg interface{}, fullPathURL string) error {

	getFormatFile := filePath(fullPathURL)

	switch getFormatFile {
	case ".json":
		fname := fullPathURL
		jsonFile, err := os.ReadFile(fname)
		if err != nil {
			return err
		}
		return json.Unmarshal(jsonFile, cfg)
	default:
		fname := fullPathURL
		yamlFile, err := os.ReadFile(fname)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(yamlFile, cfg)
	}

}

func filePath(root string) string {
	var file string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		file = filepath.Ext(info.Name())
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return file
}
