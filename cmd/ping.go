/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/blocktop/pocket-autonice/config"
	"github.com/blocktop/pocket-autonice/zeromq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Send ping messages out to all listening autonice servers.",
	Run: func(cmd *cobra.Command, args []string) {
		publisher := zeromq.NewPublisher()
		defer publisher.Close()
		topic := viper.GetString(config.PubSubTopic)
		ticker := time.NewTicker(5 * time.Second)
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		for {
			select {
			case <-sigs:
				println()
				return
			case <-ticker.C:
				if err := publisher.Publish([]byte("ping"), topic); err != nil {
					fmt.Printf("\nfailed to publish ping: %s\n", err)
				}
				fmt.Printf(".")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
