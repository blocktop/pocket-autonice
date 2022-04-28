package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Runs the client that receives notice of relays from the server.",
	Long: `The client receives a message from the serve containing the chain ID
of each relay request. The client's responsibility is to boost the "nice"
of the corresponding blockchain (if any) and maintain a timer to revert
the "niceness" after a period of no relays being received.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
