// Package cmd handles CLI commands
package cmd

/*
Copyright © 2020 Regents of the University of California

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
	"errors"
	"fmt"
	"rpm/config"
	rlog "rpm/log"
	"rpm/tycon"
	"strconv"
	"time"
)

const (
	relay1           = "1"
	relay2           = "2"
	relay3           = "3"
	relay4           = "4"
	relayCmdSet      = "set"
	relayCmdShow     = "show"
	relayCmdCycle    = "cycle"
	relayStateOpen   = "open"
	relayStateClosed = "closed"
)

var relays stringSlice
var relayCommands stringSlice
var relayStates stringSlice

func init() {
	relays = stringSlice{relay1, relay2, relay3, relay4}
	relayCommands = stringSlice{relayCmdSet, relayCmdShow, relayCmdCycle}
	relayStates = stringSlice{relayStateOpen, relayStateClosed}
}

type stringSlice []string

func (arr *stringSlice) contains(val string) bool {
	for _, elem := range *arr {
		if elem == val {
			return true
		}
	}
	return false
}

func relayParseArgs(args []string) (string, string, string, error) {

	var err error

	if len(args) < 3 {
		err = errors.New("not enough parameters, the relay command requires a relay number and an action")
		return "", "", "", err
	}

	action := args[1]
	if !relayCommands.contains(action) {
		err = fmt.Errorf("invalid relay action: %s", action)
		return "", "", "", err
	}

	relay := args[2]
	if !relays.contains(relay) {
		err = fmt.Errorf("invalid relay: %s", relay)
		return "", "", "", err
	}

	targetState := ""
	if action == relayCmdSet {
		if len(args) < 4 {
			err = errors.New("not enough parameters, the 'relay set' command requires a relay number, action and target state (open, closed)")
			return "", "", "", err
		}
		targetState = args[3]
		if !relayStates.contains(targetState) {
			err = fmt.Errorf("invalid relay state: %s", targetState)
			return "", "", "", err
		}
	}

	rlog.NoticeMsg(fmt.Sprintf("relay: %s, action: %s, targetState: %s\n", relay, action, targetState))

	return relay, action, targetState, nil
}

// Relay sets, gets, and cycles relays
func Relay(host, port string, rpmCfg *config.RPMConfig, args []string) error {

	cfg.Host = host
	cfg.Port = port
	cfg.RPMCfg = rpmCfg

	rlog.NoticeMsg(fmt.Sprintf("running %s command on host: %s:%s\n", args[0], cfg.Host, cfg.Port))

	relay, action, targetState, err := relayParseArgs(args)
	if err != nil {
		return err
	}

	initOids(cfg.RPMCfg)

	tp2din := tycon.NewTPDin2()
	err = tp2din.InitAndConnect(cfg.Host, cfg.Port)
	if err != nil {
		return err
	}
	defer tp2din.SNMPParams.Conn.Close()

	ts, results, err := tp2din.QueryOids(&relayOids)
	if err != nil {
		rlog.ErrMsg("error querying device %s:%s", cfg.Host, cfg.Port)
		return err
	}

	switch action {
	case relayCmdShow:
		displayRelayInfo(relay, ts, results)
	case relayCmdSet:
		{
			fmt.Printf("set relay %s to %s\n", relay, targetState)
		}
	case relayCmdCycle:
		{
			fmt.Printf("cycle relay %s\n", relay)
		}
	}

	return nil

}

func relayStatePretty(state string) string {
	switch state {
	case "0":
		return relayStateOpen
	case "1":
		return relayStateClosed
	default:
		return ""
	}
}

func displayRelayInfo(relay string, ts time.Time, results map[string]string) {

	fmt.Printf("%s", ts.Format("2006-01-02 15:04:05 MST"))

	for ndx, val := range cfg.RPMCfg.Oids.Relays {
		if strconv.Itoa(ndx+1) == relay {
			fmt.Printf(" relay=%s state=%s label=%s\n",
				relay,
				relayStatePretty(results[val.Oid]),
				val.Label)

		}
	}
}
