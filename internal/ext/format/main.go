package format

import "github.com/spf13/cobra"

func Commands() []*cobra.Command {
	o := &options{}

	cmd := &cobra.Command{
		Use:  "@format",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.files = args

			return o.run()
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&o.allowUnknown, "allow-unknown", false, "do not report error on unknown files")

	return []*cobra.Command{cmd}
}
