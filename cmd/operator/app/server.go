package app

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/mathspanda/ws-operator-demo/pkg/operator"
)

var (
	kubeConfig     string
	watchNamespace string
	resyncSeconds  uint32
)

var serverCmd = &cobra.Command{
	Use:           "server",
	Short:         "Launch server",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := &operator.OperatorConfig{
			KubeConfigPath: kubeConfig,
			WatchNamespace: watchNamespace,
			ResyncPeriod:   time.Duration(resyncSeconds) * time.Second,
		}

		operator, err := operator.NewOperator(config)
		if err != nil {
			return err
		}

		ctx := context.TODO()
		stopCh := make(chan struct{})

		return operator.Run(ctx, stopCh)
	},
}

func init() {
	serverCmd.Flags().StringVarP(&kubeConfig, "kubeconfig", "c", "", "path to kube config")
	serverCmd.Flags().StringVarP(&watchNamespace, "watchNamespace", "n", "",
		"namespace which operator watches")
	serverCmd.Flags().Uint32Var(&resyncSeconds, "resyncSeconds", 30,
		"resync seconds")

	rootCmd.AddCommand(serverCmd)
}
