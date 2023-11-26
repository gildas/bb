/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"bitbucket.org/gildas_cherruel/bb/cmd/pullrequest"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Log is the logger for this application
var Log *logger.Logger

// RootOptions describes the options for the application
type RootOptions struct {
	ConfigFile     string           `mapstructure:"-"`
	Bootstrap      bool             `mapstructure:"-"` // True if we are creating a config file
	LogDestination string           `mapstructure:"-"`
	ProfileName    string           `mapstructure:"-"`
	CurrentProfile *profile.Profile `mapstructure:"-"`
	Verbose        bool             `mapstructure:"-"`
	Debug          bool             `mapstructure:"-"`
}

// CmdOptions contains the options for the application
var CmdOptions RootOptions

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     APP,
	Version: Version(),
	Short:   "BitBucket Command Line Interface",
	Long: `BitBucket Command Line Interface is a tool to manage your BitBucket.
You can manage your pull requests, issues, profiles, repositories, users, etc. with it.
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	Log.Flush()
	if err != nil {
		Die(1, "Error: %s", err)
	}
}

func init() {
	Log = logger.Create(APP, &logger.NilStream{})
	configDir, err := os.UserConfigDir()
	cobra.CheckErr(err)

	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&CmdOptions.ConfigFile, "config", "", "config file (default is .env, "+filepath.Join(configDir, "bitbucket", "config-cli.yml"))
	rootCmd.PersistentFlags().StringVarP(&CmdOptions.ProfileName, "profile", "p", "", "Profile to use. Overrides TSGGLOBAL_PROFILE environment variable")
	rootCmd.PersistentFlags().StringVarP(&CmdOptions.LogDestination, "log", "l", "", "Log destination (stdout, stderr, file, none), overrides LOG_DESTINATION environment variable")
	rootCmd.PersistentFlags().BoolVar(&CmdOptions.Debug, "debug", false, "logs are written at DEBUG level, overrides DEBUG environment variable")
	rootCmd.PersistentFlags().BoolVarP(&CmdOptions.Verbose, "verbose", "v", false, "Verbose mode, overrides VERBOSE environment variable")

	rootCmd.AddCommand(profile.Command)
	rootCmd.AddCommand(pullrequest.Command)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configDir, err := os.UserConfigDir()
	cobra.CheckErr(err)

	// Bind environment variables
	_ = viper.BindEnv("DEBUG")
	_ = viper.BindEnv("VERBOSE")
	_ = viper.BindEnv("LOG_DESTINATION")
	_ = viper.BindEnv("BITBUCKET_PROFILE")

	if len(CmdOptions.ConfigFile) > 0 { // Use config file from the flag.
		viper.SetConfigFile(CmdOptions.ConfigFile)
	} else if len(configDir) > 0 {
		viper.AddConfigPath(filepath.Join(configDir, "bitbucket"))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config-cli.yml")
	} else { // Old fashion configuration file in $HOME
		homeDir, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(homeDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".bitbucket-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Read the config file
	err = viper.ReadInConfig()
	if verr, ok := err.(viper.ConfigFileNotFoundError); ok {
		Error("%s", verr)
		CmdOptions.Bootstrap = true
	} else if err != nil {
		Die(1, "Failed to read config file: %s", err)
	}

	os.Setenv("DEBUG", viper.GetString("DEBUG"))
	os.Setenv("LOG_DESTINATION", viper.GetString("LOG_DESTINATION"))
	if len(viper.GetString("LOG_DESTINATION")) == 0 {
		if viper.GetBool("DEBUG") {
			Log = logger.Create(APP, APP+".log")
		}
	} else {
		Log = logger.Create(APP, viper.GetString("LOG_DESTINATION"))
	}
	profile.Log = Log.Child("profile", "profile")
	pullrequest.Log = Log.Child("pullrequest", "pullrequest")

	Log.Infof(strings.Repeat("-", 80))
	Log.Infof("Starting %s v%s (%s)", APP, Version(), runtime.GOARCH)
	Log.Infof("Log Destination: %s", Log)
	Log.Infof("Config File: %s", viper.ConfigFileUsed())

	if err := profile.Profiles.Load(); err != nil {
		Die(1, "Failed to load profiles: %s", err)
	}
	if len(CmdOptions.ProfileName) > 0 {
		var found bool

		if CmdOptions.CurrentProfile, found = profile.Profiles.Find(CmdOptions.ProfileName); !found {
			Die(1, "Profile %s not found", CmdOptions.ProfileName)
		}
	} else {
		CmdOptions.CurrentProfile = profile.Profiles.Current()
	}
	Log.Infof("Profile: %s", CmdOptions.CurrentProfile)
	pullrequest.Profile = CmdOptions.CurrentProfile
}
