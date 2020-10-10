package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port              uint   `yaml:"port"`
	Network           string `yaml:"network"`
	RichNodeRpc       string `yaml:"rich_node_rpc"`
	Secp256k1Blake160 struct {
		Deps []struct {
			TxHash  string `yaml:"txHash"`
			Index   uint   `yaml:"index"`
			DepType string `yaml:"depType"`
		} `yaml:"deps"`
		Script struct {
			CodeHash string `yaml:"codeHash"`
			HashType string `yaml:"hashType"`
		} `yaml:"script"`
	} `yaml:"secp256k1Blake160"`
	Secp256k1Blake160Mutisig struct {
		Deps []struct {
			TxHash  string `yaml:"txHash"`
			Index   uint   `yaml:"index"`
			DepType string `yaml:"depType"`
		} `yaml:"deps"`
		Script struct {
			CodeHash string `yaml:"codeHash"`
			HashType string `yaml:"hashType"`
		} `yaml:"script"`
	} `yaml:"secp256k1Blake160Mutisig"`
	ACP struct {
		Deps []struct {
			TxHash  string `yaml:"txHash"`
			Index   uint   `yaml:"index"`
			DepType string `yaml:"depType"`
		} `yaml:"deps"`
		Script struct {
			CodeHash string `yaml:"codeHash"`
			HashType string `yaml:"hashType"`
		} `yaml:"script"`
	} `yaml:"acp"`
	UDT struct {
		Deps []struct {
			TxHash  string `yaml:"txHash"`
			Index   uint   `yaml:"index"`
			DepType string `yaml:"depType"`
		} `yaml:"deps"`
		Script struct {
			CodeHash string `yaml:"codeHash"`
			HashType string `yaml:"hashType"`
		} `yaml:"script"`
		Tokens map[string]struct {
			Symbol  string `yaml:"symbol"`
			Decimal int    `yaml:"decimal"`
		} `yaml:"tokens"`
	} `yaml:"udt"`
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
