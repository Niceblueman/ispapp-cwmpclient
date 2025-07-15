package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cwmp-client",
	Short: "A TR-069 CWMP client with UCI integration",
	Long:  `A command-line TR-069 client that communicates with an ACS and collects data from UCI using external scripts.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags()
}
