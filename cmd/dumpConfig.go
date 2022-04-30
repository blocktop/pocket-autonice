package cmd

import (
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
		if _, err := file.WriteString(configExample); err != nil {
			log.Fatalf("failed to write example config file: %s", err)
		}
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

const configExample = `# Place the config.yaml file in either the $HOME/.pocket-autonice directory
# or the /etc/pocket-autonice directory.

# All config values can be set with environment variables with precedence
# over this file by prefixing the uppercase key with "AUTONICE_".
# For example AUTONICE_LOG_LEVEL will set the log level.

# For each blockchain running on the server, add a map from the relay network
# ID or chain ID to the Linux user under with that blockchain is running. Do
# not use 'root' as a user here as the renice is done at the user level.
# Boosting the nice of root may have unintended consequences on server
# performance. NOTE: by default NO CHAINS are configured and so no processes
# would be reniced by default. Thus this configuration is mandatory. Chains can
# also be configured by setting environment variables such as
# AUTONICE_CHAINS_0021.
# chains:
#   0001: pocket  # enables pocket renice during all relayy sessions'
#   0005: fuse
#   0009: polygon
#   etc...


# Port that pocket-core prometheus is configured on. This value can be found in
# the pocket-core config.json file.
# prometheus_port: 8083

# The address to bind ZeroMQ sockets to. If pocket-core relies on relay
# blockchains on other servers over the network, then set this to the LAN IP
# address of the pocket-core server. If all blockchains are running locally,
# then this value can can be left as localhost.
# zeromq_address: 127.0.0.1:5555

# The pub/sub topic is arbitrary, but if it needs to be can be changed here.
# pubsub_topic: pocket-autonice

# When a blockchain is receiving relays, the Linux user that it is running
# under will be upgraded to this nice value. Zero is normal, negative values
# boost priority. The max boost is at -20, though that is not recommended
# as the blockchain would then compete with essential kernel services.
# nice_value: -10

# Once the blockchain stops receiving relays, the client will wait for this
# many minutes before reverting to a nice value of 0.
# nice_revert_delay_minutes: 5

# Logs will be output to this level of verbosity. Valid values are panic,
# fatal, error, warn, info, debug, and trace.
# log_level: info

# To make the logger output in JSON format, set this to true.
# log_format_json: false
`
