package local

import (
	"github.com/sirupsen/logrus"
	"github.com/fdev-ci/ci-runner/internal/pkg/common"
	"github.com/fdev-ci/ci-runner/internal/pkg/executors"
)

type LocalExecutor struct {
}

func (e *LocalExecutor) Prepare(job common.Job) error {
	logrus.Info("Prepare")
	return nil
}

func (e *LocalExecutor) Run() error {
	logrus.Info("Run")
	return nil
}

func (e *LocalExecutor) Wait() error {
	logrus.Info("Wait")
	return nil
}
func (l *LocalExecutor)SendError(err error)  {

}
func (l *LocalExecutor) Cleanup() error {
	logrus.Info("Cleanup")
	return nil
}

func createFn() common.Executor {
	return &LocalExecutor{}
}
func init()  {
	common.RegisterExecutor("local",executors.DefaultExecutorProvider{Creator: createFn})
}
