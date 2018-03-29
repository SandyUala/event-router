package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version    string
	gitCommit  string
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Event Router version",
		Long:  "Event Router version",
		Run:   printVersion,
	}
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

func printVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Event Router Version: %s\n", version)
	fmt.Printf("Git Commit: %s\n", gitCommit)
}
