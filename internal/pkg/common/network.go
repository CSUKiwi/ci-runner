package common

type VersionInfo struct {
	Name         string       `json:"name,omitempty"`
	Version      string       `json:"version,omitempty"`
	Revision     string       `json:"revision,omitempty"`
	Platform     string       `json:"platform,omitempty"`
	Architecture string       `json:"architecture,omitempty"`
	Executor     string       `json:"executor,omitempty"`
}

type JobRequest struct {
	Info       VersionInfo  `json:"info,omitempty"`
	Token      string       `json:"token,omitempty"`
}
// JobResponse 为解析 json 的结构体
type JobResponse struct {
	ID            int            `json:"id"`
	Token         string         `json:"token"`
	JobInfo       JobInfo        `json:"job_info"`
}
type JobInfo struct {
	Name        string `json:"name"`
	Timestamp   int64  `json:"timestamp"`
	Stage       string `json:"stage"`
	Image		Image  `json:"image"`
	Services    Services `json:"services,omitempty"`
	Kubernetes KubernetesConfig	`json:"kubernetes,omitempty"`
	Variables  JobVariables		`json:"variables,omitempty"`
	Timeout 	int32  `json:"timeout,omitempty"`



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


type Services []Image

type Image struct {
	Name       string   `json:"name"`
	Alias      string   `json:"alias,omitempty"`
	Command    []string `json:"command,omitempty"`
	Entrypoint []string `json:"entrypoint,omitempty"`
}