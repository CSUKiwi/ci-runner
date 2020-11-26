package executors

import (
	"github.com/tluo-github/ci-runner/internal/pkg/common"
)
type DefaultExecutorProvider struct {
	Creator          func() common.Executor
}

func (e DefaultExecutorProvider) CanCreate() bool {
	return e.Creator != nil
}

func (e DefaultExecutorProvider) Create() common.Executor {
	if e.Creator == nil {
		return nil
	}
	return e.Creator()
}
