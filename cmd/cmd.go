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
	"os"
	"os/signal"
	"rpm/config"
	"syscall"
)

// config holds parameters for the STATUS command
type cmdConfig struct {
	Host   string
	Port   string
	RPMCfg *config.RPMConfig
}

// dataOids are the OID endpoints that we will poll the device for
var dataOidInfo []config.OidInfo
var dataOids []string

// staticOids are the OID endpoints that do not change for a given device and FW version
var staticOidInfo []config.OidInfo
var staticOids []string

// relayOids are the Relay OIDs
var relayOidInfo []config.OidInfo
var relayOids []string

// allOids are the data+static OIDs
var allOidInfo []config.OidInfo
var allOids []string

var cfg cmdConfig
var sigdone chan bool

func init() {

	sigdone = setupSignals(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

}

// initialize OID vars
func initOids(c *config.RPMConfig) {

	// from the config collect dataoids to be polled
	dataOids, dataOidInfo = c.DataOidsInfo()
	staticOids, staticOidInfo = c.StaticOidsInfo()
	relayOids, relayOidInfo = c.RelayOidsInfo()
	allOids = append(staticOids, dataOids...)
	allOidInfo = append(staticOidInfo, dataOidInfo...)

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
