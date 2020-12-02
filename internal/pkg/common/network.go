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
	Volumes      []Volume	`json:"volumes,omitempty"`
	Variables  JobVariables		`json:"variables,omitempty"`
	Atoms		Atoms	`json:"atoms"`
	Timeout 	int32  `json:"timeout,omitempty"`
}

type Atoms struct {
	Count	int `json:"count"`
}
type AtomRequest struct {
	JobId	   int       `json:"jod_id"`
	Token      string    `json:"token"`	// job 的 token
	Index	   int 		 `json:"index"`  // 该job 的 atom 下标

}
type AtomResponse struct {
	ID            int            `json:"id"`
	Script		  string     	 `json:"script"`
}

type Volume struct {
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