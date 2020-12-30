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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"rpm/config"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var host string
var rpmCfg *config.RPMConfig

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
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
	//	Run: func(cmd *cobra.Command, args []string) { },
	Args: func(cmd *cobra.Command, args []string) error {

		// fmt.Printf("root func with args: %v\n", args)
		if len(args) != 2 {
			return errors.New("wrong number of commandline parameters")
		}
		return nil

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// fmt.Println("executing root")
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
	// rootCmd.PersistentFlags().StringVar(&host, "host", "", "target SNMP host")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".rpm" (without extension).
		cfgDir := filepath.Join(home, "/etc")
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
	fmt.Println("Using config file:", viper.ConfigFileUsed())

	rpmCfg = config.NewRpmCfg()
	rpmCfg.CfgFile = viper.ConfigFileUsed()
	viper.Unmarshal(&rpmCfg)

	// fmt.Printf("Checking read of config. %v\n", viper.GetStringMapString("OIDs.enterprises.45621.2.2.3.0"))
	// fmt.Printf("Checking unmarshlled cfg General value: %s\n", rpmcfg.General.Sta)
	// fmt.Printf("Checking unmarshlled cfg Relays value: %v\n", rpmcfg.OIDs.Relays[0])

	if err := rpmCfg.Validate(); err != nil {
		fmt.Printf(err.Error())
	}
}
