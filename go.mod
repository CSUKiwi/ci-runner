module github.com/tluo-github/ci-runner

go 1.14

replace github.com/tluo-github/ci-runner => ./

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/jpillora/backoff v1.0.0
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/prometheus/client_golang v0.9.3
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.0 // indirect
	github.com/stretchr/testify v1.3.0
)
