package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

func Run(file string, pValue string) error {
	value := pValue

	ext := filepath.Ext(file)

	f, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if strings.HasPrefix(pValue, "[]") {
		var result map[interface{}]interface{}

		if ext == "json" {
			if err = json.Unmarshal(f, &result); err != nil {
				return err
			}
		} else if ext == "yaml" {
			if err = yaml.Unmarshal(f, &result); err != nil {
				return err
			}
		} else {
			return errors.New("unknown file extension")
		}

		fmt.Print(result[value[2:]])
	} else if strings.HasPrefix(pValue, ".") {
		var result map[string]interface{}
		if ext == "json" {
			if err = json.Unmarshal(f, &result); err != nil {
				return err
			}
		} else if ext == "yaml" {
			if err = yaml.Unmarshal(f, &result); err != nil {
				return err
			}
		} else {
			return errors.New("unknown file extension")
		}

		fmt.Print(result[value[1:]])
	} else {
		var result map[string]interface{}
		if ext == "json" {
			if err = json.Unmarshal(f, &result); err != nil {
				return err
			}
		} else if ext == "yaml" {
			if err = yaml.Unmarshal(f, &result); err != nil {
				return err
			}
		} else {
			return errors.New("unknown file extension")
		}

		fmt.Print(result[value])
	}

	return nil
}
