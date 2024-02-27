package util

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func ReadYamlFile[T any](filename string) (T, error) {
	var t T

	data, err := os.ReadFile(filename)
	if err != nil {
		return t, fmt.Errorf("failed to load yaml file %v: %v", filename, err)
	}
	if err := yaml.Unmarshal(data, &t); err != nil {
		return t, fmt.Errorf("failed to unmarshal yaml file %v: %v", filename, err)
	}
	return t, nil
}