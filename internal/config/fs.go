package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type forDirResult struct {
	fileName string
	err      error
}

func forDir(dirName string, fn func(*os.File) error) ([]forDirResult, error) {
	result := []forDirResult{}
	dir, err := os.ReadDir(dirName)
	if err != nil {
		return result, err
	}
	for _, v := range dir {
		if v.Type().IsRegular() {
			if err := forFileRead(v.Name(), fn); err != nil {
				result = append(result, forDirResult{fileName: v.Name(), err: err})
			}
		}
	}
	return result, nil
}

func forFileRead(fileName string, fn func(*os.File) error) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0744)
	if err != nil {
		return err
	}
	defer file.Close()
	err = fn(file)
	if err != nil {
		return err
	}
	return nil
}

func decodeFile(fileName string, v any) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0744)
	if err != nil {
		return ConfigError{errType: ConfigErrorOpenFile, msg: fmt.Sprintf("config error occured; OpenFile reported: %s", err.Error())}
	}
	defer file.Close()
	dec := yaml.NewDecoder(file)
	err = dec.Decode(v)
	if err != nil {
		return ConfigError{errType: ConfigErrorDecodeError, msg: fmt.Sprintf("config error occured; yaml.decoder.decode() reported: %s", err.Error())}
	}
	return nil
}
