package main

import (
	"os"

	"kube-graph/pkg/cmd"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {

	graphCmd := cmd.NewGraphCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := graphCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
