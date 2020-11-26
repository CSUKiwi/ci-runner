package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)
type Job struct {
	Runner RunnerConfig
	JobResponse
}

func (j *Job) Run() error {
	var err error
	provider :=	GetExecutor(j.Runner.Executor)
	if provider == nil {
		err = errors.New("couldn't get provider")
		return err
	}
	executor := provider.Create()
	job_name := fmt.Sprintf("ci-runner-%s-%d",j.JobResponse.JobInfo.Name,j.JobResponse.JobInfo.Timestamp)

	json_byte, _ := json.Marshal(&j.JobResponse)
	logrus.WithFields(logrus.Fields{
		"job_name": job_name,
		"executor": j.Runner.Executor,
		"json": string(json_byte),
	}).Info("job start")


	if err == nil {
		err = executor.Prepare(*j)
	}
	if err == nil {
		err = executor.Run()
	}
	if err == nil {
		err = executor.Wait()
	} else {
		// 记录执行失败
		executor.SendError(err)
	}

	executor.Cleanup()

	logrus.WithFields(logrus.Fields{
		"job_name": job_name,
	}).Info("job end")

	return err
}


