package common

import (
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	Nacos nacosSettings `yaml:"nacos"`
}

type nacosSettings struct {
	UserName string        `yaml:"username"`
	Password string        `yaml:"password"`
	Servers  []nacosServer `yaml:"servers"`
}

type nacosServer struct {
	Host string `yaml:"host"`
	Port uint64 `yaml:"port"`
}

func NewConfig(path string) (*config, error) {

	f, err := os.ReadFile(path + "/config.yaml")
	if err != nil {
		return nil, err
	}

	var c config
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return nil, err

	}

	return &c, nil
}
