package config

import (
	"os"

	"github.com/y-kzm/go-radvd-manager/internal/radvd"
	"gopkg.in/yaml.v3"
)

const (
	filePath = "../../../configs/param.yaml"
)

func LoadDefaultParameterFile() (*radvd.Radvd, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err // ポインタではないので、ゼロ値を返す
	}

	var config radvd.Radvd
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil // ポインタではなく構造体を返す
}
