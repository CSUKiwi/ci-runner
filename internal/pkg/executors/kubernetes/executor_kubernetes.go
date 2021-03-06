package kubernetes

import (
	"bytes"
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

const (
	buildContainerName  = "build"
	helperContainerName = "helper"
)

const (
	BuildStageGetSources               string = "get_sources"
	BuildStageDownloadArtifacts        string = "download_artifacts"
	BuildStageUserScript               string = "build_script"
	BuildStageArchiveCache             string = "archive_cache"
	BuildStageUploadOnSuccessArtifacts string = "upload_artifacts_on_success"
	BuildStageUploadOnFailureArtifacts string = "upload_artifacts_on_failure"
)

type KubernetesExecutor struct {
	job           common.Job
	kubeClient    *kubernetes.Clientset
	options       *kubernetesOptions
	pod           *api.Pod
	pod_name      string
	buildFinish   chan error
	BuildLog      *os.File
	ObjectName    string
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
	e.pod_name = fmt.Sprintf("runner-%s-%d", e.job.JobInfo.JobName, e.job.JobInfo.Timestamp)

	// 创建 build log
	filename := fmt.Sprintf("/logs/%s.log", e.pod_name)
	build_log, err := os.Create(filename)

	if err != nil {
		return err
	}
	e.BuildLog = build_log

	// 连接 K8s cluster
	e.kubeClient, err = k8s_helper.GetKubeClient(e.job.Runner.Kubernetes)
	if err != nil {
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
func (e *KubernetesExecutor) prepareOptions() {
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
	apiclient := network.NewCiApiClient()

	get_source_script := "ci-runner-helper git"
	err = <-e.runInContainer(helperContainerName, get_source_script)
	if err != nil && strings.Contains(err.Error(), "command terminated with exit code") {
		return &common.BuildError{Inner: err}
	}

	//job_before_script := "ci-runner-helper artifact dowlonad"
	//
	//
	//api atom beigin
	//atom_before_script :=  "ci-runnner-helper atom stage=before jobId=luotao atomIndex=i"
	//atom_script := "sh -c goBash"
	//atom_after_script := "ci-runner-helper atom stage=after jobId=luotao atomIndex=i"
	//api  end  ok/error
	//job_after_script = "回掉api,job status ok/error"
	////循环处理 atoms
	for i := 0; i < e.job.JobInfo.Atoms.Count; i++ {
		atomData, healthy := apiclient.RequestAtom(e.job.Runner, e.job.JobInfo.Pipeline, e.job.JobInfo.StageIndex, e.job.JobInfo.JobIndex, i, e.job.Token)
		if healthy != true {
			logrus.Errorln("RqeustAtom is not healthy !")
			return errors.New("RqeustAtom is not healthy")
		}
		if atomData != nil {
			atom_before_script := fmt.Sprintf("ci-runner-helper atom_before  "+
				"--stageIndex=%d "+
				"--jobIndex=%d "+
				"--atomIndex=%d "+
				"--atomName=%s "+
				"--packagePath=%s "+
				"--inputPath=%s",
				e.job.JobInfo.StageIndex,
				e.job.JobInfo.JobIndex,
				i,
				atomData.AtomInfo.AtomName,
				atomData.AtomInfo.Execution.PackagePath,
				atomData.AtomInfo.Execution.InputPath)

			logrus.Debugln(fmt.Sprintf(
				"Starting in container %q with atom_before_script: %s",
				e.pod_name,
				atom_before_script,
			))
			err = <-e.runInContainer(helperContainerName, atom_before_script)
			if err != nil && strings.Contains(err.Error(), "command terminated with exit code") {
				return &common.BuildError{Inner: err}
			}

			//StringBuffer sb .appen()expor key=value
			sb := bytes.Buffer{}
			sb.WriteString("#!/usr/bin/env bash\n")
			sb.WriteString("set -eo pipefail\n")
			sb.WriteString("set +o noclobber\n")

			//todo 需特殊处理
			sb.WriteString("export PATH=$JAVA_HOME/bin:$PATH\n")
			sb.WriteString("export PATH=$MAVEN_HOME/bin:$PATH\n")

			for _, b := range atomData.AtomInfo.Variables {
				sb.WriteString(fmt.Sprintf("export %s=%s \n", b.Key, b.Value))
			}

			// 根据atom目录规则
			ci_data_dir := fmt.Sprintf("%s/stage-%d/job-%d/atom-%d",
				e.job.JobInfo.Variables.Get("CI_WORKSPACE"), e.job.JobInfo.StageIndex, e.job.JobInfo.JobIndex, i)
			sb.WriteString(fmt.Sprintf("export ci_data_dir=%s \n", ci_data_dir))
			sb.WriteString("export ci_data_input=input.json\n")
			sb.WriteString("export ci_data_output=output.json\n")
			sb.WriteString(fmt.Sprintf("cd %s \n", ci_data_dir))
			for _, b := range atomData.AtomInfo.Execution.Demands {
				sb.WriteString(b + "\n")
			}
			sb.WriteString(atomData.AtomInfo.Execution.Target)

			atom_build_script := sb.String()
			logrus.Debugln(fmt.Sprintf(
				"Starting in container %q with atom_build_script script: %s",
				e.pod_name,
				atom_build_script,
			))
			err = <-e.runInContainer(buildContainerName, atom_build_script)
			if err != nil && strings.Contains(err.Error(), "command terminated with exit code") {
				return &common.BuildError{Inner: err}
			}

			atom_after_script := fmt.Sprintf("ci-runner-helper atom_after "+
				"--pipeline=%s "+
				"--stageIndex=%d "+
				"--jobIndex=%d "+
				"--atomIndex=%d "+
				"--outputPath=%s",
				e.job.JobInfo.Pipeline,
				e.job.JobInfo.StageIndex,
				e.job.JobInfo.JobIndex,
				i,
				atomData.AtomInfo.Execution.OutputPath)

			err = <-e.runInContainer(helperContainerName, atom_after_script)
			if err != nil && strings.Contains(err.Error(), "command terminated with exit code") {
				return &common.BuildError{Inner: err}
			}
		}
	}

	//before_atom_git := "#!/usr/bin/env bash\nset -eo pipefail\nset +o noclobber\n\nexport ci_data_dir=/workspace/xxyyzz/stage-0/job-0/atom-0\nexport ci_data_input=input.json\nexport ci_data_output=output.json\nmkdir -p $ci_data_dir\nwget http://10.244.167.188:80/goBash -O $ci_data_dir/goBash\nwget http://172.20.10.3:8080/api/v4/atom/git/input -O $ci_data_dir/input.json\nchmod +x $ci_data_dir/goBash\nsh -c $ci_data_dir/goBash\ncat $ci_data_dir/$ci_data_output | curl -v -X POST -H \"Content-Type: application/json\" http://172.20.10.3:8080/api/v4/atom/output -d @-"
	//logrus.Debugln(fmt.Sprintf(
	//	"Starting in container %q with script: %s",
	//	e.pod_name,
	//	before_atom_git,
	//))
	//err = <-e.runInContainer(before_atom_git)
	//if err != nil && strings.Contains(err.Error(), "command terminated with exit code") {
	//	return &common.BuildError{Inner: err}
	//}
	//
	//
	//
	//script := "#!/usr/bin/env bash\nset -eo pipefail\nset +o noclobber\n\nexport PATH=$JAVA_HOME/bin:$PATH\nexport PATH=$MAVEN_HOME/bin:$PATH\n\nexport ci_data_dir=/workspace/xxyyzz/stage-0/job-0/atom-1\nexport ci_data_input=input.json\nexport ci_data_output=output.json\n\nmkdir -p $ci_data_dir\nwget http://10.244.167.188:80/goBash -O $ci_data_dir/goBash\nwget http://172.20.10.3:8080/api/v4/atom/maven/input -O $ci_data_dir/input.json\nchmod +x $ci_data_dir/goBash\nsh -c $ci_data_dir/goBash\ncat $ci_data_dir/$ci_data_output | curl -v -X POST -H \"Content-Type: application/json\" http://172.20.10.3:8080/api/v4/atom/output -d @-"
	//logrus.Debugln(fmt.Sprintf(
	//	"Starting in container %q with script: %s",
	//	e.pod_name,
	//	script,
	//))
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
	command := []string{"sh", "-c", common.BashDetectShell}
	//command := []string{"sh","-c","tail -f /dev/null"}

	build_container := e.buildContainer(buildContainerName, e.options.Image.Name)
	helper_container := e.buildContainer(helperContainerName, e.job.Runner.Kubernetes.HelperImage)
	build_container.Command = command
	helper_container.Command = command

	PodDNSConfig := api.PodDNSConfig{
		Nameservers: e.job.Runner.Kubernetes.DNS,
		Searches:    nil,
		Options:     nil,
	}

	pod_resource := &api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.pod_name,
			Namespace: e.job.Runner.Kubernetes.Namespace,
		},
		Spec: api.PodSpec{
			Volumes: e.getVolumes(),
			Containers: append([]api.Container{
				build_container, helper_container,
			}, services...),
			RestartPolicy: api.RestartPolicyNever,
			DNSConfig:     &PodDNSConfig,
		},
	}
	// 创建 pod,添加重试功能
	pod, err := e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Create(pod_resource)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"pod_name": e.pod_name,
		}).Warnln("setupBuildPod create pod [1] with error: ", err)

		time.Sleep(30 * time.Second)
		pod, err = e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Get(e.pod_name, metav1.GetOptions{})
		if err != nil {
			pod, err = e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Create(pod_resource)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"pod_name": e.pod_name,
				}).Warnln("setupBuildPod create pod [2] with error: ", err)

				time.Sleep(30 * time.Second)
				pod, err = e.kubeClient.CoreV1().Pods(e.job.Runner.Kubernetes.Namespace).Get(e.pod_name, metav1.GetOptions{})
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

func (e *KubernetesExecutor) runInContainer(name string, script string) <-chan error {
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
		command := []string{"sh", "-c", common.BashDetectShell}

		exec := ExecOptions{
			PodName:       e.pod.Name,
			Namespace:     e.pod.Namespace,
			ContainerName: name,
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
func (e *KubernetesExecutor) SendError(err error) {

}
func (e *KubernetesExecutor) Cleanup() error {
	logrus.Info("Cleanup")

	return nil
}

// 生成容器
func (e *KubernetesExecutor) buildContainer(name string, image string) api.Container {
	privileged := true

	liveness := api.Probe{
		Handler: api.Handler{
			Exec: &api.ExecAction{Command: []string{"sh", "-c", "kill me"}},
		},
		InitialDelaySeconds: e.getTimeout(),
		TimeoutSeconds:      1,
		PeriodSeconds:       1,
		SuccessThreshold:    1,
		FailureThreshold:    1,
	}
	return api.Container{
		Name:            name,
		Image:           image,
		Resources:       api.ResourceRequirements{},
		ImagePullPolicy: api.PullIfNotPresent,
		SecurityContext: &api.SecurityContext{
			Privileged: &privileged,
		},
		Env:           k8s_helper.BuildVariables(e.job.JobInfo.Variables),
		VolumeMounts:  e.getVolumeMounts(),
		Stdin:         true,
		LivenessProbe: &liveness,
		//Command: command,

	}
}

// 获得所有 VolumeMounts
func (e *KubernetesExecutor) getVolumeMounts() (mounts []api.VolumeMount) {
	for _, mount := range e.job.JobInfo.Volumes {
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
	for _, volume := range s.job.JobInfo.Volumes {
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

func (e *KubernetesExecutor) getTimeout() int32 {
	if e.job.JobInfo.Timeout != 0 {
		return e.job.JobInfo.Timeout
	}
	return common.DefaultLivessTimeout

}

func createFn() common.Executor {
	return &KubernetesExecutor{}
}
func init() {
	common.RegisterExecutor("kubernetes",
		executors.DefaultExecutorProvider{Creator: createFn})
}
