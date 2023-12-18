package get

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
	"github.com/konpyutaika/nifikop/pkg/plugin/common"
)

var getExample = `
  # view all NifiUserGroup in the current namespace
  %[1]s get

  # view NifiUserGroup foo
  %[1]s get foo
`

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
		Use:          "get [NifiUserGroup name]",
		Short:        "Get NifiUserGroup",
		Example:      fmt.Sprintf(getExample, "kubectl nifikop nifiusergroup"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.complete(c, args); err != nil {
				return err
			}
			if err := o.validate(); err != nil {
				return err
			}
			return o.run()
		},
	}

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
	if len(o.args) > 1 {
		return errors.New("either one or no arguments are allowed")
	}
	return nil
}

// run runs the get command.
func (o *options) run() error {
	list := &v1alpha1.NifiUserGroupList{}

	if o.name == "" {
		if err := o.Client.List(context.TODO(), list, &client.ListOptions{Namespace: o.UserNamespace}); err != nil {
			return fmt.Errorf("unable to list NifiUserGroup: %w", err)
		}
	} else {
		item := &v1alpha1.NifiUserGroup{}
		err := o.Client.Get(context.TODO(), client.ObjectKey{Namespace: o.UserNamespace, Name: o.name}, item)
		if err != nil && apierrors.IsNotFound(err) {
			return fmt.Errorf("NifiUserGroup %s/%s not found", o.UserNamespace, o.name)
		} else if err != nil {
			return fmt.Errorf("unable to get NifiUserGroup: %w", err)
		}
		list.Items = append(list.Items, *item)
	}

	table := newTable(o.Out)
	for _, item := range list.Items {
		data := []string{item.Namespace, item.Name}

		data = append(data, string(item.Status.Id))
		data = append(data, strconv.Itoa(int(item.Status.Version)))

		table.Append(data)
	}

	// Send output.
	table.Render()

	return nil
}

func newTable(out io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Namespace", "Name", "Id", "Version"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	return table
}
