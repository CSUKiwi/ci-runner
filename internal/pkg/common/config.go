package common

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"net"
	"strings"
)

type RunnerConfig struct {
	Name          string      `toml:"name"`
	Executor      string      `toml:"executor"`
	Concurrent    int         `toml:"concurrent"`
	Log_level     string      `toml:"log_level"`
	ListenAddress string      `toml:"listen_address,omitempty" json:"listen_address"`
	KafkaConfig   KafkaConfig `toml:"kafka"`
	Kubernetes    Kubernetes  `toml:"kubernetes"`
	MinioConfig   MinioConfig `toml:"minio"`
	Docker        Docker      `toml:"docker"`
	GitConfig     GitConfig   `toml:"git"`
	RunnerCredentials
}

type RunnerCredentials struct {
	URL         string `toml:"url" json:"url" short:"u" long:"url" env:"CI_SERVER_URL" required:"true" description:"Runner URL"`
	Token       string `toml:"token" json:"token" short:"t" long:"token" env:"CI_SERVER_TOKEN" required:"true" description:"Runner token"`
}

type Docker struct {
	DNS  		   []string        `toml:"dns,omitempty" json:"dns" long:"dns" env:"DOCKER_DNS" description:"A list of DNS servers for the container to use"`
	DockerWorkspace string        `toml:"workspace"`
	Privileged      bool          `toml:"privileged,omitzero" json:"privileged" long:"privileged" env:"DOCKER_PRIVILEGED" description:"Give extended privileges to container"`
	Volumes        []DockerVolume `toml:"volumes,omitempty" json:"volumes" long:"volumes" env:"DOCKER_VOLUMES" description:"Bind-mount a volume and create it if it doesn't exist prior to mounting. Can be specified multiple times once per mountpoint, e.g. --docker-volumes 'test0:/test0' --docker-volumes 'test1:/test1'"`
}
type DockerVolume struct {
	Name   string   `toml:"name"`
	Source string	`toml:"source"`
	Target string	`toml:"target"`
	ReadOnly bool	`toml:"read_only"`
}
type MinioConfig struct {
	Url string `toml:"url"`
	Bucket string `toml:"bucket"`
	Access_key string `toml:"access_key"`
	Secret_key string `toml:"secret_key"`
}

type GitConfig struct {
	Name string `toml:"name"`
	Password string `toml:"password"`
}

type KafkaConfig struct {
	Read   Read   `toml:"read"`
	Writer Writer `toml:"writer"`
}

type Read struct {
	BrokerList	string `toml:"brokerList"`
	Topic		string `toml:"topic"`
	GroupID       string `toml:"groupID"`
}
type Writer struct {
	BrokerList	string `toml:"brokerList"`
	Topic		string `toml:"topic"`
}

type Kubernetes struct {
	Config		string `toml:"config,omitempty"`
	Namespace	string `toml:"namespace,omitempty"`
	DNS  		[]string `toml:"dns,omitempty"`
	HelperImage string   `toml:"helper_image,omitempty"`


}

// 解析配置文件
func (config *RunnerConfig) LoadConfig(config_file string) error {
	if _, err := toml.DecodeFile(config_file, config); err != nil {
		return err
	}
	return nil
}
// 获取 ListenAddress 配置地址
func (config *RunnerConfig) GetListenAddress() (string, error) {
	address := config.ListenAddress
	if config.ListenAddress != "" {
		address = config.ListenAddress
	}
	_, port, err := net.SplitHostPort(address)
	if err != nil && !strings.Contains(err.Error(), "missing port in address") {
		return "", err
	}

	if len(port) == 0 {
		return fmt.Sprintf("%s:%d", address, DefaultMetricsServerPort), nil
	}
	return address, nil
}
func (c *RunnerCredentials) GetURL() string {
	return c.URL
}
func (c *RunnerCredentials) GetToken() string {
	return c.Token
}