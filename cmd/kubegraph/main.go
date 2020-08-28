package main

import (
	"flag"
	"kubegraph/pkg/cmd"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Set("logtostderr", "true")

	// Hide flags from --help
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name != "v" {
			pflag.Lookup(f.Name).Hidden = true
		}
	})
}

func main() {
	defer klog.Flush()
	c := cmd.NewCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := c.Execute(); err != nil {
		os.Exit(1)
	}
}
