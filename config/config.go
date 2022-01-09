package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

var Config config

type config struct {
	Server server `yaml:"server"`
	Log    Log    `yaml:"log"`
	Redis  redis  `yaml:"redis"`
	Argo   argo   `yaml:"argo"`
	Argocd argocd `yaml:"argocd"`
}

type server struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

type Log struct {
	FilePath string `yaml:"filePath"`
	Level    string `yaml:"level"`
}

type redis struct {
	Address  string `yaml:"address"`
	Db       int    `yaml:"db"`
	Password string `yaml:"password"`
}
type argo struct {
	Address            string `yaml:"address"`
	Namespace          string `yaml:"namespace"`
	ServiceAccount     string `yaml:"serviceAccount"`
	Secret             bool   `yaml:"secret"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	Http1              bool   `yaml:"http1"`
	Token              string `yaml:"token"`
}

type argocd struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Project   string `yaml:"project"`
	K8sServer string `yaml:"k8sServer"`
	RepoUrl   string `yaml:"repoUrl"`
	Revision  string `json:"revision"`
}

func InitConfig(filepath string) error {
	bs, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bs, &Config)
}
