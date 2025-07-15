package cmd

import (
	"context"

	"github.com/Niceblueman/goispappd/internal/config"
	"github.com/Niceblueman/goispappd/internal/cwmp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the CWMP client",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

		cfg, err := config.LoadConfig()
		if err != nil {
			logger.Fatalf("Failed to load config: %v", err)
		}

		client := cwmp.NewCWMPClient(cfg, logger)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := client.Initialize(ctx); err != nil {
			logger.Fatalf("Initialization failed: %v", err)
		}

		logger.Info("CWMP client initialized, running periodic informs")
		select {}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
