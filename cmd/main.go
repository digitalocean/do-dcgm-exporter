package main

import (
	"fmt"
	"os"

	"github.com/digitalocean/do-dcgm-exporter/cmd/agent"
)

func main() {
	rootCommand := agent.NewCommandStartAgent()

	if err := rootCommand.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
