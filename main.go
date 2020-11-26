package main

import (
	"github.com/tluo-github/ci-runner/cmd"
	_ "github.com/tluo-github/ci-runner/internal/pkg/executors/kubernetes"
	_ "github.com/tluo-github/ci-runner/internal/pkg/executors/local"
)

func main() {
	cmd.Execute()
}

