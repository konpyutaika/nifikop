package nifiuser

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/konpyutaika/nifikop/cmd/kubectl-nifikop/nifiuser/get"
)

// options provides information required by clusteragent command.
type options struct {
	genericclioptions.IOStreams
	configFlags *genericclioptions.ConfigFlags
}

// newOptions provides an instance of options with default values.
func newOptions(streams genericclioptions.IOStreams) *options {
	return &options{
		configFlags: genericclioptions.NewConfigFlags(false),
		IOStreams:   streams,
	}
}

// New provides a cobra command wrapping options for "clusteragent" sub command.
func New(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use: "nifiuser [subcommand] [flags]",
	}

	cmd.AddCommand(get.New(streams))

	o := newOptions(streams)
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}
