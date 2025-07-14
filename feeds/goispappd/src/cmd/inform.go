package cmd

import (
	"github.com/Niceblueman/goispappd/internal/config"
	"github.com/Niceblueman/goispappd/internal/cwmp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var informCmd = &cobra.Command{
	Use:   "inform [event]",
	Short: "Send an Inform message to the ACS",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

		cfg, err := config.LoadConfig()
		if err != nil {
			logger.Fatalf("Failed to load config: %v", err)
		}

		client := cwmp.NewCWMPClient(cfg, logger)
		if err := client.SendInform(args[0]); err != nil {
			logger.Fatalf("Failed to send inform: %v", err)
		}
		logger.Info("Inform sent successfully")
	},
}

func init() {
	rootCmd.AddCommand(informCmd)
}
