package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "ws-operator-demo",
	Short: "ws-operator-demo is an demo operator for web server cluster",
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Please use -h to see usage")
	},
}

var debugLevel uint32

func init() {
	rootCmd.PersistentFlags().Uint32VarP(&debugLevel, "debuglevel", "l", 4,
		"log debug level: 0[panic] 1[fatal] 2[error] 3[warn] 4[info] 5[debug]")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.Level(debugLevel))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
