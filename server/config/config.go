package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port        uint   `yaml:"port"`
	Network     string `yaml:"network"`
	RichNodeRpc string `yaml:"rich_node_rpc"`
}

func Init(path string) (*Config, error) {
	var c Config

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
