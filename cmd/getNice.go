/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/blocktop/pocket-autonice/renicer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// getNiceCmd represents the getNice command
var getNiceCmd = &cobra.Command{
	Use:   "get-nice <chainID>",
	Short: "Gets the current niceness of the Linux user associated with the given chainID",
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
		nice, err := renicer.GetNiceValue(chainID)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		user := viper.GetString(chainID)
		fmt.Printf("Chain %s (%s) niceness: %d\n", chainID, user, nice)
	},
}

func init() {
	rootCmd.AddCommand(getNiceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getNiceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getNiceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
