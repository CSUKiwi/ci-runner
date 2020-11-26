package run

import (
	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/tluo-github/ci-runner/internal/pkg/common"
	"github.com/tluo-github/ci-runner/internal/pkg/network"
	"net"
	"net/http"
	"os"
	"time"
)


var (
	config common.RunnerConfig 		// ci runner 配置文件
	currentWorkers int				// 当前运行的job并发数

)
// 设置日志
func initLog()  {
	file, error := os.OpenFile("/logs/ci-runner.log", os.O_CREATE|os.O_WRONLY |os.O_APPEND, 0666)
	if error == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
	}
	logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.999"})
	logrus.SetLevel(logrus.DebugLevel)
}
// 加载配置文件
func initConfig(cfgFile string)  {
	if _, err := toml.DecodeFile(cfgFile, &config); err != nil {
		logrus.WithError(err).Fatal("Failed to load config file")
	}
}
// 配置 prometheus metrics
func initMetricsAndDebugServer()  {
	listenAddress, err := config.GetListenAddress()
	if err != nil {
		logrus.Errorf("invalid listen address: %s", err.Error())
		return
	}

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create listener for metrics server")
	}

	mux := http.NewServeMux()
	serveMetrics(mux)
	go func() {
		err := http.Serve(listener, mux)
		if err != nil {
			logrus.WithError(err).Fatal("Metrics server terminated")
		}
	}()

	logrus.
		WithField("address", listenAddress).
		Info("Metrics server listening")
}

func serveMetrics(mux *http.ServeMux) {
	registry := prometheus.NewRegistry()
	// 暴露 程序版本
	//registry.MustRegister(versionCollector)

	// 暴露关于进程的Go-specific指标(GC stats, goroutines，等等)。
	registry.MustRegister(prometheus.NewGoCollector())
	// 暴露与go无关的进程指标(内存使用、文件描述符等)。
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

}
// ci-runner run 命令入口函数
func Run(cfgFile string)  {
	logrus.Info("CI runner Run")
	initConfig(cfgFile)
	initLog()
	initMetricsAndDebugServer()
	run()
}
// ci runner 业务 run 入口函数
func run()  {
	apiclient :=network.NewCiApiClient()
	// 未到并发上线,每隔 3s 获取 job
	for {
		if currentWorkers <= config.Concurrent{
			jobData, healthy := apiclient.RequestJob(config)
			if healthy == false {
				logrus.Errorln("Runner is not healthy and will be disabled!")
			}
			if jobData != nil {
				go func() {
					new_job := common.Job{
						Runner: config,
						JobResponse: *jobData,
					}
					new_job.Run()
				}()

			}
		}
		time.Sleep(1000 * time.Second)
	}
}




