package cmd

import (
	"context"
	"github.com/blocktop/pocket-autonice/client"
	"github.com/blocktop/pocket-autonice/config"
	"github.com/blocktop/pocket-autonice/prometheusPoller"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "Runs the autonice service.",
	Long: `The autonice service receives a message from the poller via zeromq containing
the chain ID of each relay request. The service boosts the "nice" of the
corresponding blockchain (if any) and maintains a timer to revert the
"niceness" after a period of no relays being received. The service must
be run on all blockchain nodes where this functionality is desired.

On the pocket-core node, the service must run the poller as well via the
--with-poller flag.

Use the --dry-run flag verify communication between poller and blockchain
servers without changing the nice value of any process.
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.InitConfig)
	clientCmd.Flags().BoolVar(&poller, "with-poller", false,
		`starts the client and the prometheus poller
(use on the server that is running pocket-core)`)
	clientCmd.Flags().BoolVar(&dryRun, "dry-run", false,
		`runs all functionality except renicing processes`)
}
