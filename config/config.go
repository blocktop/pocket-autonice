package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

const (
	LogLevel               = "log_level"
	LogFormatJSON          = "log_format_json"
	NiceValue              = "nice_value"
	NiceRevertDelayMinutes = "nice_revert_delay_minutes"
	SubscriberAddress      = "subscriber_address"
	PublishToEndpoints     = "publisher_endpoints"
	PubSubTopic            = "pubsub_topic"
	PrometheusPort         = "prometheus_port"
	RunWithSudo            = "run_with_sudo"
	Chains                 = "chains"
)

// InitConfig initializes the configuration for the CLI. See documentation.
// Use the dump-config command to generate a config.yaml file.
func InitConfig() {

	viper.SetDefault(LogLevel, "info")
	viper.SetDefault(LogFormatJSON, false)
	viper.SetDefault(NiceValue, -10)
	viper.SetDefault(NiceRevertDelayMinutes, 5)
	viper.SetDefault(SubscriberAddress, "127.0.0.1:5555")
	viper.SetDefault(PublishToEndpoints, []string{"127.0.0.1:5555"})
	viper.SetDefault(PubSubTopic, "pocket-autonice")
	viper.SetDefault(PrometheusPort, 8083)
	viper.SetDefault(RunWithSudo, false)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/pocket-autonice")
	viper.AddConfigPath("$HOME/.pocket-autonice")

	viper.SetEnvPrefix("AUTONICE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error reading config file: %w \n", err))
		}
	}

	var level log.Level
	switch strings.ToLower(viper.GetString(LogLevel)) {
	case "panic":
		level = log.PanicLevel
	case "fatal":
		level = log.FatalLevel
	case "error":
		level = log.ErrorLevel
	case "warn", "warning":
		level = log.WarnLevel
	case "debug":
		level = log.DebugLevel
	case "trace":
		level = log.TraceLevel
	default:
		level = log.InfoLevel
	}
	log.SetLevel(level)

	logJson := viper.GetBool(LogFormatJSON)
	var logFormatter log.Formatter
	if logJson {
		customFormatter := &log.JSONFormatter{}
		logFormatter = customFormatter
		customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	} else {
		customFormatter := &log.TextFormatter{}
		logFormatter = customFormatter
		customFormatter.TimestampFormat = "2006-01-02 15:04:05"
		customFormatter.FullTimestamp = true
		customFormatter.PadLevelText = true
	}
	log.SetFormatter(logFormatter)
}

const ConfigExample = `# Place the config.yaml file in either the $HOME/.pocket-autonice directory
# or the /etc/pocket-autonice directory.

# All config values can be set with environment variables with precedence
# over this file by prefixing the uppercase key with "AUTONICE_".
# For example AUTONICE_LOG_LEVEL will set the log level.

# Autonice runs the Linux 'renice' command which is privileged. The program must
# either be run as root or with sudo.
# run_with_sudo: false

# For each blockchain running on the server, add a map from the relay network
# ID or chain ID to the Linux user under with that blockchain is running. Do
# not use 'root' as a user here as the renice is done at the user level.
# Boosting the nice of root may have unintended consequences on server
# performance. NOTE: by default NO CHAINS are configured and so no processes
# would be reniced by default. Thus this configuration is mandatory.
# chains:
#   "0001": pocket  # enables pocket renice during all relay sessions'
#   "0005": fuse
#   "0009": polygon
#   etc...


# Port that pocket-core prometheus is configured on. This value can be found in
# the pocket-core config.json file.
# prometheus_port: 8083

# For the client, the address to bind ZeroMQ subscriber socket. If pocket-core
# relies on a network relay blockchains on other servers over a LAN, then set
# this to the LAN IP address of the client node. Note that the pocket-core
# node should also be setup as a client to receive messages to it. If all
# blockchains are running locally, then this value can can be left as
# localhost (the default).
# subscriber_address: 127.0.0.1:5555

# For the server, the addresses of all client sockets in the network that
# the local pocket-core relies on to server relays. If pocket-core and
# blockchains are all running locally then set only one entry here to
# localhost (the default).
# publish_to_endpoints:
#   - 127.0.0.1:5555

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
