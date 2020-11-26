module github.com/fdev-ci/ci-runner

go 1.14

replace github.com/fdev-ci/ci-runner => ./

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jpillora/backoff v1.0.0
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/prometheus/client_golang v0.9.3
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.0 // indirect
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.0.0-20190620084959-7cf5895f2711
	k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719
	k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/utils v0.0.0-20200731180307-f00132d28269 // indirect
)
