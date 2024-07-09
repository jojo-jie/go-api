package configs

import (
	"embed"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

//go:embed config.yaml
var configs embed.FS

type Config struct {
	Minio MinioConfig `yaml:"minio"`
}

type MinioConfig struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	Api        string `yaml:"api"`
	Path       string `yaml:"path"`
	BucketName string `yaml:"bucket_name"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) Init() error {
	configStr, err := configs.ReadFile("config.yaml")
	if os.IsNotExist(err) {
		return err
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(configStr)))
	err = decoder.Decode(c)
	if err != nil {
		return err
	}
	return nil
}
