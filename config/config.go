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

func InitConfig() {
	// set chain config in environment variables to identify which Linux user each chain is running under
	// export AUTONICE_0009=polygon
	// export AUTONICE_0021=geth
	// etc..

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
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error reading config file: %w \n", err))
		}
	}
	viper.SetEnvPrefix("AUTONICE")
	viper.AutomaticEnv()

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
