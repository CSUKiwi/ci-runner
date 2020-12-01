package kubernetes

import (
	"errors"
	"fmt"
	"github.com/fdev-ci/ci-runner/internal/pkg/common"
	"github.com/fdev-ci/ci-runner/internal/pkg/executors"
	"github.com/fdev-ci/ci-runner/internal/pkg/network"
	k8s_helper "github.com/fdev-ci/ci-runner/pkg/helpers/k8s"
	"github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"strings"
	"time"
)
type KubernetesExecutor struct {
	job common.Job
	kubeClient  *kubernetes.Clientset
	options     *kubernetesOptions
	pod         *api.Pod
	pod_name    string
	buildFinish chan error
	BuildLog    *os.File
	ObjectName  string
	IsSystemError bool
}

type kubernetesOptions struct {
	Image    common.Image    //build image
	Services common.Services // services image
}


func (e *KubernetesExecutor) Prepare(job common.Job) error {
	logrus.Info("Prepare")
	e.buildFinish = make(chan error, 1)
	e.IsSystemError = false
	e.job = job
	e.pod_name = fmt.Sprintf("runner-%s-%d",e.job.JobInfo.Name,e.job.JobInfo.Timestamp)

	// 创建 build log
	filename := fmt.Sprintf("/logs/%s.log",e.pod_name)
	build_log, err := os.Create(filename)

	if err != nil {
		return err
	}
	e.BuildLog = build_log

	// 连接 K8s cluster
	e.kubeClient, err = k8s_helper.GetKubeClient(e.job.Runner.Kubernetes)
	if err != nil{
		logrus.WithFields(logrus.Fields{
			"pod_name": e.pod_name,
		}).Errorln("connection k8s faild with error:", err)
		e.IsSystemError = true
		return errors.New("connection k8s faild with error")
	}
	// 处理 options config 相关
	e.prepareOptions()

	// k8s pod 初始化
	err = e.setupBuildPod()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"pod_name": e.pod_name,
		}).Errorln("setupBuildPod with error: ", err)
		e.IsSystemError = true
		return err
	}
	return nil
}
func (e *KubernetesExecutor) prepareOptions()  {
	e.options = &kubernetesOptions{}
	e.options.Image = e.job.JobInfo.Image
	for _, service := range e.job.JobInfo.Services {
		if service.Name == "" {
			continue
		}
		e.options.Services = append(e.options.Services, service)
	}
}




func (e *KubernetesExecutor) Run() error {
	logrus.Info("Run")
	var err error
	// 开始通过 k8s client 让 pod 执行命令
	apiclient :=network.NewCiApiClient()

	// 循环处理 atoms
	for i := 0; i < 2; i++ {
		atomData, healthy := apiclient.RequestAtom(e.job.Runner,e.job.ID,e.job.Token,i)
		if healthy != true{
			logrus.Errorln("RqeustAtom is not healthy !")
			return errors.New("RqeustAtom is not healthy")
		}
		if atomData != nil {
			logrus.Debugln(fmt.Sprintf(
				"Starting in container %q with script: %s",
				e.pod_name,
				atomData.Script,
			))
				err = <-e.runInContainer(atomData.Script)
			if err != nil && strings.Contains(err.Error(), "command terminated with exit code") {
				return &common.BuildError{Inner: err}
			}
		}
	}





	//script := "#!/usr/bin/env bash\n\nset -eo pipefail\nset +o noclobber\nexport ci_data_dir=/tmp/1/workspace\nexport ci_data_input=input.json\nexport ci_data_output=output.json\nmkdir -p $ci_data_dir\nwget http://172.20.10.3:8080/api/v4/atom/input -O $ci_data_dir/input.json\nwget http://10.107.250.219:80/goDemo -O $ci_data_dir/goDemo\nchmod +x $ci_data_dir/goDemo\nsh -c $ci_data_dir/goDemo\ncat $ci_data_dir/$ci_data_output | curl -v -X POST -H \"Content-Type: application/json\" http://172.20.10.3:8080/api/v4/atom/output -d @-"
	//logrus.Debugln(fmt.Sprintf(
	//	"Starting in container %q with script: %s",
	//	e.pod_name,
	//	script,
	//))
	//err := <-e.runInContainer(script)
	//if err != nil && strings.Contains(err.Error(), "command terminated with exit code") {
	//	return &common.BuildError{Inner: err}
	//}
	//
	//
	//script = "#!/usr/bin/env bash\n\nset -eo pipefail\nset +o noclobber\nexport ci_data_dir=/tmp/2/workspace\nexport ci_data_input=input.json\nexport ci_data_output=output.json\nmkdir -p $ci_data_dir\nwget http://172.20.10.3:8080/api/v4/atom/gobash/input -O $ci_data_dir/input.json\nwget http://10.107.250.219:80/goBash -O $ci_data_dir/goBash\nchmod +x $ci_data_dir/goBash\nsh -c $ci_data_dir/goBash\ncat $ci_data_dir/$ci_data_output | curl -v -X POST -H \"Content-Type: application/json\" http://172.20.10.3:8080/api/v4/atom/output -d @-"
	//logrus.Debugln(fmt.Sprintf(
	//	"Starting in container %q with script: %s",
	//	e.pod_name,
	//	script,
	//))
	//
	//err = <-e.runInContainer(script)
	//if err != nil && strings.Contains(err.Error(), "command terminated with exit code") {
	//	return &common.BuildError{Inner: err}
	//}

	return err


}


// 初始化设置 pod
func (e *KubernetesExecutor) setupBuildPod() error {

	logrus.WithFields(logrus.Fields{
		"pod_name": e.pod_name,
	}).Info("setupBuildPod")

	services := make([]api.Container, len(e.options.Services))
	for i, service := range e.options.Services {
		services[i] = e.buildContainer(fmt.Sprintf("svc-%d", i), service.Name)
	}
	//todo step labels
	//todo step annotations
	//todo step imagePullSecrets
	command := []string{"sh","-c",common.BashDetectShell}
	//command := []string{"sh","-c","tail -f /dev/null"}
	pod_container := e.buildContainer("build", e.options.Image.Name)
	pod_container.Command = command
	PodDNSConfig := api.PodDNSConfig{
		Nameservers: e.job.Runner.Kubernetes.DNS,
		Searches:    nil,
		Options:     nil,
	}

	pod_resource := &api.Pod{
		ObjectMeta : metav1.ObjectMeta{
			Name: e.pod_name,
			Namespace: e.job.Runner.Kubernetes.Namespace,
		},
		Spec: api.PodSpec{
			Volumes: e.getVolumes(),
			Containers: append([]api.Container{
				pod_container,
			}, services...),
			RestartPolicy: api.RestartPolicyNever,
			DNSConfig:  &PodDNSConfig,
		},
	}
	// 创建 pod,添加重试功能
	pod, err := e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Create(pod_resource)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"pod_name": e.pod_name,
		}).Warnln("setupBuildPod create pod [1] with error: ", err)

		time.Sleep(30 * time.Second)
		pod, err =e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Get(e.pod_name,metav1.GetOptions{})
		if err != nil {
			pod, err = e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Create(pod_resource)
			if err != nil{
				logrus.WithFields(logrus.Fields{
					"pod_name": e.pod_name,
				}).Warnln("setupBuildPod create pod [2] with error: ", err)

				time.Sleep(30 * time.Second)
				pod, err =e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Get(e.pod_name,metav1.GetOptions{})
				if err != nil {
					pod, err = e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Create(pod_resource)
					if err != nil {
						return err
					}
				}

			}
		}

	}
	e.pod = pod
	return nil

}


func (e *KubernetesExecutor) runInContainer(script string) <-chan error  {
	errc := make(chan error, 1)

	go func() {
		status, err := k8s_helper.WaitForPodRunning(e.kubeClient, e.pod)
		if err != nil {
			e.buildFinish <- err
			return
		}
		if status != api.PodRunning {
			e.IsSystemError = true
			e.buildFinish <- fmt.Errorf("pod failed to enter running state: %s", status)
			return
		}

		logrus.WithFields(logrus.Fields{
			"pod_name": e.pod_name,
		}).Info("pod state is PodRunning")

		config, err := k8s_helper.GetKubeClientConfig(e.job.Runner.Kubernetes)
		if err != nil {
			e.buildFinish <- err
			return
		}
		command := []string{"sh","-c",common.BashDetectShell}

		exec := ExecOptions{
			PodName:       e.pod.Name,
			Namespace:     e.pod.Namespace,
			ContainerName: "build",
			Command:       command,
			In:            strings.NewReader(script),
			Out:           e.BuildLog,
			Err:           e.BuildLog,
			Stdin:         true,
			Config:        config,
			Client:        e.kubeClient,
			Executor:      &DefaultRemoteExecutor{},
		}
		errc <- exec.Run()
	}()
	return errc

}

func (e *KubernetesExecutor) Wait() error {
	logrus.Info("Wait")
	return nil
}
func (e *KubernetesExecutor)SendError(err error)  {

}
func (e *KubernetesExecutor) Cleanup() error {
	logrus.Info("Cleanup")
	return nil
}

// 生成容器
func (e *KubernetesExecutor) buildContainer(name string, image string)  api.Container {
	privileged := true





	liveness := api.Probe{
		Handler:  api.Handler{
			Exec: &api.ExecAction{Command: []string{"sh","-c","kill me"}},
		},
		InitialDelaySeconds: e.getTimeout(),
		TimeoutSeconds:      1,
		PeriodSeconds:       1,
		SuccessThreshold:    1,
		FailureThreshold:    1,
	}
	return api.Container{
		Name:                     name,
		Image:                    image,
		Resources:                api.ResourceRequirements{},
		ImagePullPolicy:          api.PullIfNotPresent,
		SecurityContext:          &api.SecurityContext{
			Privileged: &privileged ,
		},
		Env:             k8s_helper.BuildVariables(e.job.JobInfo.Variables),
		VolumeMounts:			  e.getVolumeMounts(),
		Stdin:                    true,
		LivenessProbe: &liveness,
		//Command: command,

	}
}

// 获得所有 VolumeMounts
func (e *KubernetesExecutor) getVolumeMounts() (mounts []api.VolumeMount) {
	for _, mount := range e.job.JobInfo.Kubernetes.Volumes.Host_paths {
		mounts = append(mounts, api.VolumeMount{
			Name:      mount.Name,
			MountPath: mount.Mount_path,
			ReadOnly:  mount.Read_only,
		})
	}
	return
}

// 获得所有的 Volume
func (s *KubernetesExecutor) getVolumes() (volumes []api.Volume) {
	for _, volume := range s.job.JobInfo.Kubernetes.Volumes.Host_paths {
		path := volume.Host_path
		// Make backward compatible with syntax introduced in version 9.3.0
		if path == "" {
			path = volume.Host_path
		}

		volumes = append(volumes, api.Volume{
			Name: volume.Name,
			VolumeSource: api.VolumeSource{
				HostPath: &api.HostPathVolumeSource{
					Path: path,
				},
			},
		})
	}
	return
}

func (e *KubernetesExecutor) getTimeout ()int32 {
	if e.job.JobInfo.Timeout != 0{
		return e.job.JobInfo.Timeout
	}
	return common.DefaultLivessTimeout

}


func createFn() common.Executor {
	return &KubernetesExecutor{}
}
func init()  {
	common.RegisterExecutor("kubernetes",
		executors.DefaultExecutorProvider{Creator: createFn})
}
