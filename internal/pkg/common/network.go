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
type JobResponse struct {
	ID            int            `json:"id"`
	Token         string         `json:"token"`
}
