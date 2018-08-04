package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/gsmcwhirter/eso-discord/pkg/options"
)

func setup(start func(config) error) *options.Command {
	cli := options.NewCLI(AppName, BuildVersion, BuildSHA, BuildDate, options.CommandOptions{
		ShortHelp: "Manage the discord bot",
		Args:      options.NoArgs,
	})

	var configFile string

	cli.Flags().StringVar(&configFile, "config", "./config.toml", "The config file to use")
	cli.Flags().String("user", "", "The discord user string to impersonate")
	cli.Flags().String("database", "", "The database file")
	cli.Flags().String("test_thing", "", "Testing")

	cli.SetRunFunc(func(cmd *options.Command, args []string) (err error) {
		v := viper.New()

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

		return start(c)
	})

	return cli
}
