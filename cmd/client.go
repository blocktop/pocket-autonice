package cmd

import (
	"context"
	"github.com/blocktop/pocket-autonice/client"
	"github.com/blocktop/pocket-autonice/prometheusPoller"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var poller bool
var dryRun bool
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Runs the client that receives notice of relays from the poller.",
	Long: `The client receives a message from the poller via zeromq containing
the chain ID of each relay request. The client's responsibility is to boost the
"nice" of the corresponding blockchain (if any) and maintain a timer to revert
the "niceness" after a period of no relays being received.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		viper.Set("dry_run", dryRun)
		if poller {
			go prometheusPoller.Start(ctx)
		}
		go client.Start(ctx)

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().BoolVar(&poller, "withPoller", false, `starts the client and the prometheus poller
(use on the server that is running pocket-core)`)
	clientCmd.Flags().BoolVar(&dryRun, "dry-run", false, `runs all functionality except renicing processes`)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
