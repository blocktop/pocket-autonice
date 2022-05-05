package cmd

import (
	"fmt"
	"github.com/blocktop/pocket-autonice/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path"
)

// dumpConfigCmd represents the dumpConfig command
var dumpConfigCmd = &cobra.Command{
	Use:   "dump-config",
	Short: "Dumps a defalt configuration file at $HOME/.pocket-autonice/config_example.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir := os.Getenv("HOME")
		configDir := path.Join(homeDir, ".pocket-autonice")
		configFile := path.Join(configDir, "config_example.yaml")

		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("failed to create config directory %s: %s", configDir, err)
		}

		if _, err := os.Stat(configFile); !os.IsNotExist(err) {
			if err = os.Remove(configFile); err != nil {
				log.Fatalf("failed to remove existing %s: %s", configFile, err)
			}
		}

		file, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("unable to open config file for writing: %s", err)
		}
		defer file.Close()
		if _, err := file.WriteString(config.ConfigExample); err != nil {
			log.Fatalf("failed to write example config file: %s", err)
		}
		fmt.Printf("Example config generated at: %s\n", configFile)
	},
}

func init() {
	rootCmd.AddCommand(dumpConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
