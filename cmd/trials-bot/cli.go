package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/gsmcwhirter/eso-discord/pkg/options"
)

func setup(start func(config) error) *options.Command {
	cli := options.NewCLI(AppName, BuildVersion, BuildSHA, BuildDate, options.CommandOptions{
		ShortHelp:    "Manage the discord bot",
		Args:         options.NoArgs,
		SilenceUsage: true,
	})

	var configFile string

	cli.Flags().StringVar(&configFile, "config", "./config.toml", "The config file to use")
	cli.Flags().String("client_id", "", "The discord bot client id")
	cli.Flags().String("client_secret", "", "The discord bot client secret")
	cli.Flags().String("client_token", "", "The discord bot client token")
	cli.Flags().String("database", "", "The database file")
	cli.Flags().String("log_format", "", "The logger format")
	cli.Flags().String("log_level", "", "The minimum log level to show")
	cli.Flags().Int("num_workers", 0, "The number of worker goroutines to run")
	cli.Flags().String("pprof_hostport", "", "The host and port for the pprof http server to listen on")

	cli.SetRunFunc(func(cmd *options.Command, args []string) (err error) {
		v := viper.New()

		v.SetDefault("pprof_hostport", "127.0.0.1:6060")

		if configFile != "" {
			v.SetConfigFile(configFile)
		} else {
			v.SetConfigName("config")
			v.AddConfigPath(".") // working directory
		}

		v.SetEnvPrefix("EDB")
		v.AutomaticEnv()

		err = v.BindPFlags(cmd.Flags())
		if err != nil {
			return errors.Wrap(err, "could not bind flags to viper")
		}

		err = v.ReadInConfig()
		if err != nil {
			return errors.Wrap(err, "could not read in config file")
		}

		c := config{}
		err = v.Unmarshal(&c)
		if err != nil {
			return errors.Wrap(err, "could not unmarshal config into struct")
		}

		c.Version = cmd.Version

		return start(c)
	})

	return cli
}
