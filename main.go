package main

/*
Copyright © 2020 Regents of the University of California

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

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"rpm/cmd"
	"rpm/config"
	rlog "rpm/log"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	defaultUser string = "nrts"
)

type ctxKeyType string

type appConfig struct {
	debug   bool
	cfgFile string
	cmd     string
	host    string
	port    string
	rpmCfg  *config.RPMConfig
}

var appCfg = &appConfig{
	debug:   false,
	cfgFile: "",
	cmd:     "",
	host:    "",
	port:    "",
	rpmCfg:  nil,
}

func initLogging(debug bool) error {

	severityLevel := syslog.LOG_NOTICE
	if debug {
		severityLevel = syslog.LOG_DEBUG
	}
	err := rlog.InitLogging("rpm", syslog.LOG_LOCAL0|severityLevel)
	if err != nil {
		return err
	}
	if appCfg.debug {
		rlog.SetLogLevel(syslog.LOG_DEBUG)
	}
	return nil
}

func logStartup() error {

	abspath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return err
	}
	rlog.NoticeMsg(fmt.Sprintf("%s starting up...", abspath))
	return nil
}

func setUser(username string) error {

	var uid, gid int
	nrtsuser, err := user.Lookup(username)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		rlog.WarningMsg(err.Error())
	} else {
		if uid, err = strconv.Atoi(nrtsuser.Uid); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			rlog.ErrMsg("could not convert uid %v to int", uid)
			rlog.ErrMsg(err.Error())
		}
		if gid, err = strconv.Atoi(nrtsuser.Gid); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			rlog.ErrMsg("could not convert gid %v to int", gid)
			rlog.ErrMsg(err.Error())
		}
		syscall.Setuid(uid)
		syscall.Setgid(gid)
		os.Chdir(nrtsuser.HomeDir)

		var wd string
		wd, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			rlog.ErrMsg("could not determine working dir", gid)
			rlog.ErrMsg(err.Error())
		} else {
			rlog.NoticeMsg(fmt.Sprintf("working dir: %s", wd))
		}
	}
	return err
}

func main() {

	var err error

	initFlags()

	err = initLogging(appCfg.debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, "rpm: error creating logger (fprintf)")
		log.Fatal("rpm: error creating logger (log.fatal)")
	}

	err = readCLI()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		usage()
		flag.PrintDefaults()
		rlog.ErrMsg(err.Error())
		os.Exit(1)
	}

	err = logStartup()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error(), "here")
		rlog.ErrMsg(err.Error())
		os.Exit(1)
	}

	err = setUser(defaultUser)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		rlog.ErrMsg(err.Error())
		os.Exit(1)
	}

	// read rpm config file
	appCfg.rpmCfg, err = initTPDin2Config(appCfg.cfgFile)

	executeCmd()

	rlog.NoticeMsg("%s shutting down", os.Args[0])
}

func executeCmd() {

	var err error

	switch appCfg.cmd {
	case "poll":
		err = cmd.Poll(appCfg.host, appCfg.port, appCfg.rpmCfg, os.Args[2:])
	case "status":
		err = cmd.Status(appCfg.host, appCfg.port, appCfg.rpmCfg, os.Args[2:])
	case "relay":
		err = cmd.Relay(appCfg.host, appCfg.port, appCfg.rpmCfg, os.Args[2:])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		rlog.ErrMsg(err.Error())
	}
}

func initFlags() {

	flag.BoolVar(&appCfg.debug, "debug", false, "enable debug logging")
	flag.StringVar(&appCfg.cfgFile, "config", "", "specify config file")

}

func validCmd(cmd string) bool {
	validCommands := []string{
		"poll",
		"status",
		"ctl",
	}
	for _, n := range validCommands {
		if cmd == n {
			return true
		}
	}
	return false
}

// read CLI flags adjust app config appropriately
func readCLI() error {

	var err error

	flag.Parse()

	// sanity check on params; needs at least host and cmd
	if len(os.Args) < 2 {
		err := errors.New("command line error, not enough parameters")
		return err
	}

	// parse host[]:port]
	hostport := os.Args[1]
	appCfg.host, appCfg.port, err = formatSNMPHostPort(hostport)
	if err != nil {
		err = fmt.Errorf("could not parse host[:port]: %s", hostport)
		return err
	}

	// get command
	cmd := os.Args[2]
	if !validCmd(cmd) {
		err = fmt.Errorf("unrecognized command: %s", cmd)
		return err
	}
	appCfg.cmd = cmd

	return err
}

func usage() {
	fmt.Println("do it this way, dummy.")
}

// initConfig reads in config file and ENV variables if set.
func initTPDin2Config(rpmCfgFile string) (*config.RPMConfig, error) {

	var err error

	if rpmCfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(rpmCfgFile)
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

	rpmCfg := config.NewConfig()
	viper.Unmarshal(&rpmCfg)
	rpmCfg.CfgFile = viper.ConfigFileUsed()

	// fmt.Printf("Checking read of config. %v\n", viper.GetStringMapString("OIDs.enterprises.45621.2.2.3.0"))
	// fmt.Printf("Checking unmarshlled cfg General value: %s\n", rpmcfg.General.Sta)
	// fmt.Printf("Checking unmarshlled cfg Relays value: %v\n", rpmcfg.OIDs.Relays[0])

	if err := rpmCfg.Validate(); err != nil {
		fmt.Printf(err.Error())
	}

	return rpmCfg, err
}

func formatSNMPHostPort(rawHost string) (string, string, error) {

	if strings.Index(rawHost, ":") == -1 {
		rawHost += ":161"
	}

    h, p, _ := net.SplitHostPort(rawHost)
    ips, err := net.LookupHost(h)
    if err != nil {
        return "", "", err
    }

	return ips[0], p, nil

}
