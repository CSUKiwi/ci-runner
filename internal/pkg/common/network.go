package common

type VersionInfo struct {
	Name         string `json:"name,omitempty"`
	Version      string `json:"version,omitempty"`
	Revision     string `json:"revision,omitempty"`
	Platform     string `json:"platform,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Executor     string `json:"executor,omitempty"`
}

type JobRequest struct {
	Info  VersionInfo `json:"info,omitempty"`
	Token string      `json:"token,omitempty"`
}

// JobResponse 为解析 json 的结构体
type JobResponse struct {
	ID      int     `json:"id"`
	Token   string  `json:"token"`
	JobInfo JobInfo `json:"job_info"`
}
type JobInfo struct {
	Pipeline   string       `json:"pepeline"`
	Stage      string       `json:"stage"`
	StageIndex int          `json:"stage_index"`
	JobName    string       `json:"job_name"`
	JobIndex   int          `json:"job_index"`
	Timestamp  int64        `json:"timestamp"`
	Image      Image        `json:"image"`
	Services   Services     `json:"services,omitempty"`
	Volumes    []Volume     `json:"volumes,omitempty"`
	Variables  JobVariables `json:"variables,omitempty"`
	Atoms      Atoms        `json:"atoms"`
	Timeout    int32        `json:"timeout,omitempty"`
}

type Atoms struct {
	Count int `json:"count"`
}

type AtomRequest struct {
	Pipeline   string `json:"pipeline"`
	StageIndex int    `json:"stage_index"`
	JobIndex   int    `json:"job_index"`
	AtomIndex  int    `json:"atom_index"` // 该job 的 atom 下标
	Token      string `json:"job_token"` // job 的 token

}
type AtomResponse struct {
	ID     int    `json:"id"`
	Token string `json:"token"`
	AtomInfo AtomInfo `json:"atom_info"`
}
type AtomInfo struct {
	AtomName string `json:"atom_name"`
	Execution Execution  `json:"execution"`
	Variables  JobVariables `json:"variables,omitempty"`
}

type Execution struct {
	Language string `json:"language"`
	PackagePath string `json:"package_path"`
	InputPath string `json:"input_path"`
	OutputPath string `json:"output_path"`
	Demands []string `json:"demands"`
	Target string `json:"target"`
}

type Volume struct {
	Name       string `json:"name"`
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
