package stop

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/konpyutaika/nifikop/api/v1"
	"github.com/konpyutaika/nifikop/pkg/plugin/common"
	nifiutil "github.com/konpyutaika/nifikop/pkg/util/nifi"
)

var (
	del         bool
	stopExample = `
  # set label %[2]s to force the stop of NifiDataflow foo
  %[1]s stop foo
	
  # remove label %[2]s that forces the stop of NifiDataflow foo
  %[1]s stop foo
`
)

// options provides information required by Datadog get command.
type options struct {
	genericclioptions.IOStreams
	common.Options
	args []string
	name string
}

// newOptions provides an instance of getOptions with default values.
func newOptions(streams genericclioptions.IOStreams) *options {
	o := &options{
		IOStreams: streams,
	}
	o.SetConfigFlags()
	return o
}

// New provides a cobra command wrapping options for "get" sub command.
func New(streams genericclioptions.IOStreams) *cobra.Command {
	o := newOptions(streams)
	cmd := &cobra.Command{
		Use:          "stop [NifiDataflow name]",
		Short:        fmt.Sprintf("Set label %s to true on NifiDataflow", nifiutil.ForceStopLabel),
		Example:      fmt.Sprintf(stopExample, "kubectl nifikop nifidataflow", nifiutil.ForceStopLabel),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.complete(c, args); err != nil {
				return err
			}
			if err := o.validate(); err != nil {
				return err
			}
			return o.run(c)
		},
	}

	cmd.Flags().BoolVarP(&del, "delete", "d", false, fmt.Sprintf("Delete label %s on NifiDataflow", nifiutil.ForceStopLabel))
	o.ConfigFlags.AddFlags(cmd.Flags())

	return cmd
}

// complete sets all information required for processing the command.
func (o *options) complete(cmd *cobra.Command, args []string) error {
	o.args = args
	if len(args) > 0 {
		o.name = args[0]
	}
	return o.Init(cmd)
}

// validate ensures that all required arguments and flag values are provided.
func (o *options) validate() error {
	if len(o.args) != 1 {
		return errors.New("one argument must provided")
	}
	return nil
}

// run runs the stop command.
func (o *options) run(cmd *cobra.Command) error {
	item := &v1.NifiDataflow{}
	err := o.Client.Get(context.TODO(), client.ObjectKey{Namespace: o.UserNamespace, Name: o.name}, item)
	if err != nil && apierrors.IsNotFound(err) {
		return fmt.Errorf("NifiDataflow %s/%s not found", o.UserNamespace, o.name)
	} else if err != nil {
		return fmt.Errorf("unable to get NifiDataflow: %w", err)
	}

	itemOriginal := item.DeepCopy()
	labels := item.GetLabels()

	if !del {
		labels[nifiutil.ForceStopLabel] = "true"
	} else {
		delete(labels, nifiutil.ForceStopLabel)
	}

	item.SetLabels(labels)
	err = o.Client.Patch(context.TODO(), item, client.MergeFrom(itemOriginal))

	if err != nil {
		cmd.Println(fmt.Sprintf("Couldn't patch %s/%s: %v", item.GetNamespace(), item.GetName(), err))
	} else {
		cmd.Println(fmt.Sprintf("NifiDataflow labels patched successfully in %s/%s", item.GetNamespace(), item.GetName()))
	}

	return nil
}
