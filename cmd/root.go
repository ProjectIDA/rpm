// Package cmd for CLI commandline
/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log/syslog"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"

	"rpm/config"
	rlog "rpm/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	debug   bool
	host    string
	rpmCfg  *config.RPMConfig

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "rpm",
		Short: "IDA Remote Power Monitor (RPM)",
		Long: `The IDA Remote Power Monitor (RPM) polls the 
Tycon TPDin2-Web power monitor using the SNMP protocol for 
real-time power metrics. It polls the user specified teraget device
at the specified time interval (in seconds).
   The details of which mettrics are queried are controlled by a 
configuration file which is assumed to be located at $NRTS_HOME or
can be specified on the command line.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) {},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rpm.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "log debug messages")
	// rootCmd.PersistentFlags().StringVar(&host, "host", "", "target SNMP host")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// set debug level, if specified
	if debug {
		rlog.SetLogLevel(syslog.LOG_DEBUG)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		// get user home dir
		// Find home directory.
		user, err := user.Current()
		if err != nil {
			rlog.ErrMsg(err.Error())
			os.Exit(1)
		}
		rlog.NoticeMsg(fmt.Sprintf("homedir: %s", user.HomeDir))

		// Search config in home directory with name ".rpm" (without extension).
		cfgDir := filepath.Join(user.HomeDir, "/etc")
		viper.AddConfigPath(cfgDir)
		viper.AddConfigPath(".")
		viper.SetConfigName("rpm")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}
	rlog.NoticeMsg(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))

	rpmCfg = config.NewConfig()
	viper.Unmarshal(&rpmCfg)
	rpmCfg.CfgFile = viper.ConfigFileUsed()

	// fmt.Printf("Checking read of config. %v\n", viper.GetStringMapString("OIDs.enterprises.45621.2.2.3.0"))
	// fmt.Printf("Checking unmarshlled cfg General value: %s\n", rpmcfg.General.Sta)
	// fmt.Printf("Checking unmarshlled cfg Relays value: %v\n", rpmcfg.OIDs.Relays[0])

	if err := rpmCfg.Validate(); err != nil {
		fmt.Printf(err.Error())
	}
}

// SetupSignals to trap for external kill signals
func setupSignals(sigs ...os.Signal) chan bool {

	sigchan := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigchan, sigs...)

	go func() {
		sig := <-sigchan
		fmt.Println(sig)
		done <- true
	}()

	return done
}

// split host and port, supplying default SNMP port if absent
func formatSNMPHostPort(rawHost string) (string, string) {

	if strings.Index(rawHost, ":") == -1 {
		rawHost += ":161"
	}
	parts := strings.Split(rawHost, ":")

	return parts[0], parts[1]

}
