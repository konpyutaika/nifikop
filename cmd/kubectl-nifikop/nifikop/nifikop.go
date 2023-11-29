package nifikop

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/konpyutaika/nifikop/cmd/kubectl-nifikop/nificluster"
	"github.com/konpyutaika/nifikop/cmd/kubectl-nifikop/nificonnection"
	"github.com/konpyutaika/nifikop/cmd/kubectl-nifikop/nifidataflow"
	"github.com/konpyutaika/nifikop/cmd/kubectl-nifikop/nifigroupautoscaler"
	"github.com/konpyutaika/nifikop/cmd/kubectl-nifikop/nifiregistryclient"
	"github.com/konpyutaika/nifikop/cmd/kubectl-nifikop/nifiuser"
	"github.com/konpyutaika/nifikop/cmd/kubectl-nifikop/nifiusergroup"
)

// options provides information required by datadog command.
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

// NewCmd provides a cobra command wrapping options for "datadog" command.
func NewCmd(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use: "nifikop [subcommand] [flags]",
	}

	// Operator commands
	cmd.AddCommand(nificluster.New(streams))
	cmd.AddCommand(nifidataflow.New(streams))
	cmd.AddCommand(nificonnection.New(streams))
	cmd.AddCommand(nifiuser.New(streams))
	cmd.AddCommand(nifiusergroup.New(streams))
	cmd.AddCommand(nifiregistryclient.New(streams))
	cmd.AddCommand(nifigroupautoscaler.New(streams))

	o := newOptions(streams)
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}
