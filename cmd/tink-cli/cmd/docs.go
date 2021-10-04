package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsPath string

// docsCmd returns the generate command that, when run, generates
// documentation.
func docsCmd(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "docs [markdown|man]",
		Short:     "Generate documentation",
		Hidden:    true,
		ValidArgs: []string{"markdown", "man"},
		Args:      cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format := args[0]

			switch format {
			case "markdown":
				return doc.GenMarkdownTree(cmd.Parent(), docsPath)
			case "man":
				header := &doc.GenManHeader{Title: name}
				return doc.GenManTree(cmd.Parent(), header, docsPath)
			}
			// ValidArgs make this error response dead-code
			return fmt.Errorf("unknown format: %q", format)
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.Flags().StringVarP(&docsPath, "path", "p", "", "Path where documentation will be generated")
	return cmd
}

func init() {
	docsCmd := docsCmd(rootCmd.CalledAs())

	rootCmd.AddCommand(docsCmd)
}
