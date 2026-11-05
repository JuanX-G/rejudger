package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Result of running for dir. Filename and the corresponding error.
type forDirResult struct {
	FileName string
	Err      error
}

// Run the function fn on every file in the specified directory.
func forDir(dirName string, fn func(*os.File) error) ([]forDirResult, error) {
	result := []forDirResult{}
	dir, err := os.ReadDir(dirName)
	if err != nil {
		return result, err
	}
	for _, v := range dir {
		if v.Type().IsRegular() {
			if err := forFileRead(filepath.Join(dirName, v.Name()), fn); err != nil {
				result = append(result, forDirResult{FileName: v.Name(), Err: err})
			}
		}
	}
	return result, nil
}

// Open the file and apply fn to the returned to os.File.
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

func decodeOpenFile(file *os.File, v any) error {
	dec := yaml.NewDecoder(file)
	err := dec.Decode(v)
	if err != nil {
		return ConfigError{errType: ConfigErrorDecodeError, msg: fmt.Sprintf("config error occured; yaml.decoder.decode() reported: %s", err.Error())}
	}
	return nil
}
