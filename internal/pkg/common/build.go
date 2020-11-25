package common

import "time"

// Build 为解析 kafka消息的Json 结构体
type Build struct {
	Name     string				`json:"name"`
	Timeout int32				`json:"timeout"`
	Timestamp   int64  			`json:"timestamp"`
	Command string				`json:"command"`
	Image	Image 				`json:"image,omitempty"`
	Services Services			`json:"services,omitempty"`
	Kubernetes KubernetesConfig	`json:"kubernetes,omitempty"`
	SshConfig []SshConfig 		`json:"sshs,omitempty"`
	Variables  JobVariables		`json:"variables,omitempty"`
	MetaData map[string]interface{}`json:"metadata,omitempty"`
	CreatedAt time.Time
}

type Services []Image

type Image struct {
	Name       string   `json:"name"`
	Alias      string   `json:"alias,omitempty"`
	Command    []string `json:"command,omitempty"`
	Entrypoint []string `json:"entrypoint,omitempty"`
}

type SshConfig struct {
	User         string `toml:"user,omitempty" json:"user" long:"user" env:"SSH_USER" description:"User name"`
	Password     string `toml:"password,omitempty" json:"password" long:"password" env:"SSH_PASSWORD" description:"User password"`
	Host         string `toml:"host,omitempty" json:"host" long:"host" env:"SSH_HOST" description:"Remote host"`
	Port         string `toml:"port,omitempty" json:"port" long:"port" env:"SSH_PORT" description:"Remote host port"`
	IdentityFile string `toml:"identity_file,omitempty" json:"identity_file" long:"identity-file" env:"SSH_IDENTITY_FILE" description:"Identity file to be used"`
}

type KubernetesConfig struct {
	Volumes      KubernetesVolumes	`json:"volumes,omitempty"`
}
type KubernetesVolumes struct {
	Host_paths  []KubernetesHostPath  `json:"host_paths,omitempty"`
}
type KubernetesHostPath struct {
	Name      string `json:"name"`
	Mount_path string `json:"mount_path"`
	Read_only  bool   `json:"read_only"`
	Host_path  string `json:"host_path"`
}



func (b *Build) Duration() time.Duration {
	return time.Since(b.CreatedAt)
}
