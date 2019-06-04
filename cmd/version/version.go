package version

import (
	"fmt"
	"github.com/spf13/cobra"
)

// NewVersionCmd is version command line subcommand
func NewVersionCmd(version string, build string) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "A build version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %v (built: %v)\n", version, build)
		},
	}
	return cmd
}
