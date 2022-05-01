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
	ZeroMQAddress          = "zeromq_address"
	PubSubTopic            = "pubsub_topic"
	PrometheusPort         = "prometheus_port"
)

// InitConfig initializes the configuration for the CLI. See documentation.
// Use the dump-config command to generate a config.yaml file.
func InitConfig() {

	viper.SetDefault(LogLevel, "info")
	viper.SetDefault(LogFormatJSON, false)
	viper.SetDefault(NiceValue, -10)
	viper.SetDefault(NiceRevertDelayMinutes, 5)
	viper.SetDefault(ZeroMQAddress, "127.0.0.1:5555")
	viper.SetDefault(PubSubTopic, "pocket-autonice")
	viper.SetDefault(PrometheusPort, 8083)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/pocket-autonice")
	viper.AddConfigPath("$HOME/.pocket-autonice")

	viper.SetEnvPrefix("AUTONICE")
	viper.AutomaticEnv()

	var readFromFile bool
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error reading config file: %w \n", err))
		}
	} else {
		readFromFile = true
	}

	if readFromFile {
		chains := viper.GetStringMapString("chains")
		for chainID, user := range chains {
			viper.Set(chainID, user)
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

# For each blockchain running on the server, add a map from the relay network
# ID or chain ID to the Linux user under with that blockchain is running. Do
# not use 'root' as a user here as the renice is done at the user level.
# Boosting the nice of root may have unintended consequences on server
# performance. NOTE: by default NO CHAINS are configured and so no processes
# would be reniced by default. Thus this configuration is mandatory. Chains can
# also be configured by setting environment variables such as
# AUTONICE_CHAINS_0021.
# chains:
#   "0001": pocket  # enables pocket renice during all relay sessions'
#   "0005": fuse
#   "0009": polygon
#   etc...


# Port that pocket-core prometheus is configured on. This value can be found in
# the pocket-core config.json file.
# prometheus_port: 8083

# The address to bind ZeroMQ sockets to. If pocket-core relies on relay
# blockchains on other servers over the network, then set this to the LAN IP
# address of the pocket-core server. If all blockchains are running locally,
# then this value can can be left as localhost.
# zeromq_address: 127.0.0.1:5555

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
