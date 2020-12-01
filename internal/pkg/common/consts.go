package common

import "time"

const NAME = "super-runner"
const VERSION = "2020-12-01"
const DefaultMetricsServerPort = 9252	//prometheus metrics 端口
const DefaultLivessTimeout = 1800		// pod 生产时间
const KubernetesPollInterval = 10 * time.Second // 每次查询 pod 详情 间隔时间
const KubernetesPollAttempts = 100 // 尝试获得 Pod 详情次数
const SshConnectRetries = 3 // ssh 重试次数
const SshRetryInterval = 3	//ssh 间隔时间
const DefaultDockerWorkspace= "/workspace" // docker executor worspace dir
const DefaultNetworkClientTimeout = 60 * time.Minute // http 底层 timeout 超时时间

const BashDetectShell = `
if [ -x /usr/local/bin/bash ]; then
	exec /usr/local/bin/bash 
elif [ -x /usr/bin/bash ]; then
	exec /usr/bin/bash 
elif [ -x /bin/bash ]; then
	exec /bin/bash 
elif [ -x /usr/local/bin/sh ]; then
	exec /usr/local/bin/sh 
elif [ -x /usr/bin/sh ]; then
	exec /usr/bin/sh 
elif [ -x /bin/sh ]; then
	exec /bin/sh 
elif [ -x /busybox/sh ]; then
	exec /busybox/sh 
else
	echo shell not found
	exit 1
fi

`



