package cmd

import (
	"fmt"
	"kube-graph/pkg/client"
	"kube-graph/pkg/graph"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // support for cloud providers auth
)

// Options defines the graph command options
type Options struct {
	ConfigFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
	Client   *client.Client
	Name     string
	Type     string
	DotGraph bool
}

// NewOptions returns an Options struct
func NewOptions(iostreams genericclioptions.IOStreams) *Options {
	return &Options{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   iostreams,
	}
}

// NewGraphCmd returns a graph command
func NewGraphCmd(iostreams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(iostreams)

	c := &cobra.Command{
		Use:   "graph [TYPE] [NAME] [flags]",
		Short: "Print or create a graph to visualize the relation between kubernetes objects",
		Example: `
# Print a graph that shows all kubernetes objects that are related to the service service-foo
kubectl graph service service-foo
		
# Create a DOT graph that shows all kubernetes objects that are related to the ingress ingress-bar
kubectl graph ingress ingress-bar --dot-graph
`,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	c.Flags().BoolVar(&o.DotGraph, "dot-graph", o.DotGraph, "If true, a DOT graph file will be created instead of printing to stdout")
	o.ConfigFlags.AddFlags(c.Flags())

	return c
}

// Complete adds values to Option attributes
func (o *Options) Complete(cmd *cobra.Command, args []string) error {
	// Get Type and Name
	if len(args) >= 2 {
		o.Type = args[0]
		o.Name = args[1]
	}

	// Get namespace
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	if namespace == "" {
		clientConfig := o.ConfigFlags.ToRawKubeConfigLoader()

		namespace, _, err = clientConfig.Namespace()
		if err != nil {
			return err
		}
	}

	// Get Client
	restConfig, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	o.Client = client.NewClient(clientset, namespace)

	return nil
}

// Validate ensures all expected data is available
func (o *Options) Validate(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("requires TYPE and NAME arguments. Run: kubectl graph -h")
	}

	return nil
}

// Run creates a new resource and starts the data gathering to build the graph
func (o *Options) Run() error {

	b := graph.NewBuilder(o.Client, o.Type, o.Name)

	err := b.Build()
	if err != nil {
		return err
	}

	return nil
}
