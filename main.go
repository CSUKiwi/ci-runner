package main

import (
	"github.com/fdev-ci/ci-runner/cmd"
	_ "github.com/fdev-ci/ci-runner/internal/pkg/executors/kubernetes"
	_ "github.com/fdev-ci/ci-runner/internal/pkg/executors/local"
)

func main() {
	cmd.Execute()
}

