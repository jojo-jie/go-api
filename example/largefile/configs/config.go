package configs

import (
	"embed"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"
)

//go:embed config.yaml
var configs embed.FS

type Config struct {
	Minio  MinioConfig `yaml:"minio"`
	Server Server      `yaml:"server"`
}

type Server struct {
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type MinioConfig struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	Api        string `yaml:"api"`
	Path       string `yaml:"path"`
	BucketName string `yaml:"bucket_name"`
	PutExpires string `yaml:"put_expires"`
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
