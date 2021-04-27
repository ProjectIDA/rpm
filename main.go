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
        rlog.NoticeMsg(fmt.Sprintf("nrtsuser.HomeDir: %s", nrtsuser.HomeDir))

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

	// parse host[]:port]
	hostport := flag.Args()[0]
	appCfg.host, appCfg.port, err = formatSNMPHostPort(hostport)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		rlog.ErrMsg(err.Error())
		os.Exit(1)
	}

	err = readCLI(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		rlog.ErrMsg(err.Error())
		usage()
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
    if err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        rlog.ErrMsg(err.Error())
        os.Exit(1)
    }

	executeCmd(flag.Args())

	rlog.NoticeMsg("%s shutting down", os.Args[0])
}

func executeCmd(parms []string) {

	var err error

	switch appCfg.cmd {
	case "poll":
		err = cmd.Poll(appCfg.host, appCfg.port, appCfg.rpmCfg, parms[1:])
	case "status":
		err = cmd.Status(appCfg.host, appCfg.port, appCfg.rpmCfg, parms[1:])
	case "relay":
		err = cmd.Relay(appCfg.host, appCfg.port, appCfg.rpmCfg, parms[1:])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		rlog.ErrMsg(err.Error())
	}
}

func initFlags() {

	flag.BoolVar(&appCfg.debug, "debug", false, "enable debug logging")
	// flag.StringVar(&appCfg.cfgFile, "config", "", "specify config file")
	flag.Parse()

}

func validCmd(cmd string) bool {
	validCommands := []string{
		"poll",
		"status",
		"relay",
	}
	for _, n := range validCommands {
		if cmd == n {
			return true
		}
	}
	return false
}

// read CLI flags adjust app config appropriately
func readCLI(parms []string) error {

	var err error

	// sanity check on params; needs at least host and cmd
	if len(parms) < 2 {
		err := errors.New("command line error, not enough parameters")
		return err
	}

	// get command
	cmd := parms[1]
	if !validCmd(cmd) {
		err = fmt.Errorf("invalid command: %s", cmd)
		return err
	}
	appCfg.cmd = cmd

	return err
}

func usage() {
	usagesMsg := `
usage: rpm <hostname-or-ip[:port]> <command> [ command-parameters ]

Commands:
    status                - display Tycon TPDin2 current values

    poll <interval-secs>  - will poll TPDin device repeatedly, 
                            outputing results to stdout in
                            txtoida10 version 2 format

    relay <sub-command>, where <sub-sommand> is one of:
	
        show  <relay-#>                   - to show current state of relay
        cycle <relay-#>                   - to cycle relay
        set   <relay-#> { open | closed } - set relay to a new state

Examples:
    rpm 192.168.1.25 status        
    rpm 192.168.1.25 poll 1
    rpm 192.168.1.25 relay cycle 2 
    rpm 192.168.1.25 relay show 2 
    rpm 192.168.1.25 relay set 3 closed  
	`
	fmt.Println(usagesMsg)
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
		user, err := user.Lookup(defaultUser)
		if err != nil {
			rlog.ErrMsg(err.Error())
			os.Exit(1)
		}
		rlog.NoticeMsg(fmt.Sprintf("homedir: %s", user.HomeDir))

		// Search for config cwd first, then in home directory with name "rpm.toml".
		viper.AddConfigPath(".")
		cfgDir := filepath.Join(user.HomeDir, "/etc")
		viper.AddConfigPath(cfgDir)
		viper.SetConfigName("rpm")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		return nil, err
	}
	rlog.NoticeMsg(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))

	rpmCfg := config.NewConfig()
	viper.Unmarshal(&rpmCfg)
	rpmCfg.CfgFile = viper.ConfigFileUsed()

	if err := rpmCfg.Validate(); err != nil {
		fmt.Printf(err.Error())
		return nil, err
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
