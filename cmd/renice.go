package cmd

import (
	"fmt"
	"github.com/blocktop/pocket-autonice/zeromq"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

// reniceCmd represents the publish command
var reniceCmd = &cobra.Command{
	Use:   "renice",
	Short: "Publish a message to ZeroMQ, typically to manually trigger renicing of a blockchain.",
	Long: `Sometimes it may be useful to manually renice a blockchain's processes in the cluster.
The publish command provides a way to tigger this by sending a chain ID as a
message to all subscribers in the cluster. Each blockchain process will renice
to the nice_value configured on that node. For example:

autonice renice 0040

will renice all harmony processes across the cluster.'
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("exactl 1 argument is required; %d given", len(args))
		}
		if len(args[0]) != 4 {
			return fmt.Errorf("argument must be a valid 4-character pocket relay chainID")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		chainID := args[0]
		publisher, err := zeromq.NewPublisher()
		if err != nil {
			fmt.Printf("failed to start publisher: %s\n", err)
			os.Exit(1)
		}
		defer publisher.Close()
		err = publisher.Publish(chainID, chainID)
		if err != nil {
			log.Fatalf("failed to publish renice for chain %s: %s", chainID, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(reniceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reniceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reniceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
