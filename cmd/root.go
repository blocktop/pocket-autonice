package cmd

import (
	"github.com/blocktop/pocket-autonice/config"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "autonice",
	Short: "Boost pocket and blockchains when they are serving relays.",
	Long: `Pocket autonice uses the Linux operating system's "nice" feature
to boost the CPU priority of the pocket process and the blockchain
process when they are serving relays.

There are two parts to this project. A prometheus poller that monitors the
prometheus output of pocket-core for relay requests. And a client process
receives messages from the poller when relay requests are received.
These messages identify which blockchain is the target of the relays
so that its "niceness" can be boosted.

Both parts of the process run in the same process, but the poller is not
required on blockchain nodes that are not running pocket-core. So for
blockchain nodes start just the client:

   autonice client

and for nodes that are running pocket-core as well as optionally running
blockchain nodes start the client and the poller:

   autonice client --withPoller
`,
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
}
