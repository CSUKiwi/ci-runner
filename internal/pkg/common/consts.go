package common

import "time"

const NAME = "super-runner"
const VERSION = "2020-11-25"
const DefaultMetricsServerPort = 9252	//prometheus metrics 端口
const DefaultLivessTimeout = 1800		// pod 生产时间
const KubernetesPollInterval = 10 * time.Second // 每次查询 pod 详情 间隔时间
const KubernetesPollAttempts = 100 // 尝试获得 Pod 详情次数
const SshConnectRetries = 3 // ssh 重试次数
const SshRetryInterval = 3	//ssh 间隔时间
const DefaultDockerWorkspace= "/workspace" // docker executor worspace dir
const DefaultNetworkClientTimeout = 60 * time.Minute // http 底层 timeout 超时时间




