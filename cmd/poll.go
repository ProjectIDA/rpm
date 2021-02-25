// Package cmd for CLI commandline
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
package cmd

import (
	"context"
	"fmt"
	"log"
	"rpm/config"
	rlog "rpm/log"
	"rpm/tycon"
	"strconv"
	"sync"
	"time"
	// "github.com/spf13/cobra"
)

// dataOids are the OID endpoints that we will poll the device for
var dataOidInfo []config.OidInfo
var dataOids []string

// staticOids are the OID endpoints that do not change for a given device and FW version
var staticOidInfo []config.OidInfo
var staticOids []string

// allOids are the data+static OIDs
var allOidInfo []config.OidInfo
var allOids []string

// pollCmd represents the poll command
// var pollCmd = &cobra.Command{
// 	Use:   "poll host_and_port polling_interval_in_secs",
// 	Short: "Poll SNMP target for values",
// 	Long:  `Poll SNMP target for values at a fixed interval >= 1.0 seconds`,
// 	Args:  checkPollArgs, //cobra.ExactArgs(2),
// 	Run:   poll,
// }

func getSampleInterval(intstr string) (float64, error) {

	val, err := strconv.ParseFloat(intstr, 32)
	if err != nil {
		return 0, err
	}

	if (val < tycon.MinSampleInterval.Seconds()) || (val > tycon.MaxSampleInterval.Seconds()) {
		rlog.ErrMsg(fmt.Sprintf("invalid sample interval %f must be between %.0f and %.0f seconds",
			val,
			tycon.MinSampleInterval.Seconds(),
			tycon.MaxSampleInterval.Seconds()))
		return 0, fmt.Errorf("invalid sample interval %f must be between %.0f and %.0f seconds",
			val,
			tycon.MinSampleInterval.Seconds(),
			tycon.MaxSampleInterval.Seconds())
	}

	return val, nil
}

// PollingConfig holds parameters for POLL command
// type PollingConfig struct {
// 	Host         string
// 	Port         string
// 	IntervalSecs time.Duration
// 	rpmCfg       *config.RPMConfig
// }

func formatScan(sampleInterval time.Duration, cfg *config.RPMConfig, scan *tycon.TPDin2Scan) string {

	outstr := fmt.Sprintf(
		"%04d %02d %02d %02d %02d %02d",
		scan.TS.Year(),
		scan.TS.Month(),
		scan.TS.Day(),
		scan.TS.Hour(),
		scan.TS.Minute(),
		scan.TS.Second(),
	)

	outstr += fmt.Sprintf(" %s %s %s %.0f",
		cfg.General.Net,
		cfg.General.Sta,
		cfg.General.Loc,
		sampleInterval.Seconds(),
	)

	for _, oidinfo := range dataOidInfo {
		oid := oidinfo.Oid
		outstr += fmt.Sprintf(" %s:%s", oidinfo.Chancode, scan.Data[oid])
	}

	return outstr

}

// initialize OID vars
func initOids(c *config.RPMConfig) {

	// from the config collect dataoids to be polled
	dataOids, dataOidInfo = c.DataOidsInfo()
	staticOids, staticOidInfo = c.StaticOidsInfo()
	allOids = append(staticOids, dataOids...)
	allOidInfo = append(staticOidInfo, dataOidInfo...)

}

func logDeviceInfo(scan *tycon.TPDin2Scan) {

	for _, oidinfo := range staticOidInfo {
		rlog.NoticeMsg("%s: %s", oidinfo.Label, scan.Data[oidinfo.Oid])
	}

}

// Poll the TPDin2 device
func Poll(host, port string, rpmCfg *config.RPMConfig, pollArgs []string) {
	// snmpwalk -On -c readwrite -M /usr/local/share/snmp/mibs -v 1 localhost

	// rlog.DebugMsg(fmt.Sprintf("poll cmd with args[]: %v\n", args))
	var err error

	cfg.Host = host
	cfg.Port = port
	cfg.RPMCfg = rpmCfg

	intervalSecsf64, err := getSampleInterval(pollArgs[0])

	fInterval := float32(intervalSecsf64)
	dInterval := time.Duration(fInterval) * time.Second
	hInterval := dInterval / 2

	rlog.NoticeMsg(fmt.Sprintf("Host: %s:%s; interval: %.0f sec(s)\n", host, port, fInterval))

	initOids(cfg.RPMCfg)
	// sigdone := setupSignals(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	tp2din := tycon.NewTPDin2()
	err = tp2din.Initialize(host, port, dInterval)
	if err != nil {
		rlog.ErrMsg("unknown error initializing tp2din... quitting")
		log.Fatalln(err)
	}
	err = tp2din.Connect()
	if err != nil {
		rlog.CritMsg("could not connect to %s:%s... quitting.", host, port)
		log.Fatalln(fmt.Errorf("could not connect to %s:%s... quitting", host, port))
	}
	defer tp2din.SNMPParams.Conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	err = tp2din.PollStart(ctx, &wg, &allOids)
	if err != nil {
		rlog.ErrMsg("could not start internal polling loop... quitting")
		cancel()
		wg.Wait()
		log.Fatal(err)
	}
	rlog.NoticeMsg("internal polling loop spawned")

	var scan, prevScan *tycon.TPDin2Scan
	targetTime := time.Now().Round(dInterval).Add(dInterval)

	var offset time.Duration
	first := true
	scanMissed := false
	scanRepeated := false
	exiting := false

	for !exiting {

		targetTime = targetTime.Add(dInterval)
		rlog.DebugMsg("next target time: %v\n", targetTime.String())

		select {

		case <-time.After(time.Until(targetTime)):

			prevScan = scan
			scan, err = tp2din.GetScan()
			if scan == nil {
				if !scanMissed {
					rlog.ErrMsg("no rpm scan available\n")
				}
				scanMissed = true
				first = true
				continue
			}

			rlog.DebugMsg("Scan time:   %s", scan.TS.String())
			for _, oidinfo := range dataOidInfo {
				rlog.DebugMsg("(%s) %s: %s", oidinfo.Chancode, oidinfo.Oid, scan.Data[oidinfo.Oid])
			}

			if first {
				rlog.NoticeMsg("initial rpm scan received")
				logDeviceInfo(scan)
				first = false
			}
			scanMissed = false

		case <-sigdone:
			rlog.DebugMsg("got done signal")
			exiting = true
			continue
		}

		// calcualte offset of scan time from target time.
		// positive offset means scan time is after target time
		offset = scan.TS.Sub(targetTime)

		// see if scan should go with next second, perhaps due to network lag

		if offset > hInterval {
			// really should never get here unless this loop is taking more than an interval to complete
			rlog.WarningMsg("well, this is awkward, a scan from more than 1/2 interval in the future")
			rlog.WarningMsg("incrementing TargetTime by one interval (to catch up) creating a gap")
			targetTime = scan.TS.Round(dInterval) //targetTime.Add(dInterval)
			first = true
		} else if offset < -hInterval {
			// current scan does not appear to be available.
			rlog.ErrMsg("missing scan: current scan time (%v) not found within 1/2 interval of target (%v)", scan.TS, targetTime)

			// if previous scan not alreadcy repeated, repeat previous scan (if it exists)
			// and set flag so con only do this one time in a row.
			if (prevScan != nil) && (!scanRepeated) {
				rlog.WarningMsg("repeating previous scan value")
				scanRepeated = true
				scan = prevScan
			} else {
				// missed scan but can't repeat previous, so there will be a gap
				first = true
			}
			continue
		}

		scanRepeated = false
		// send record to Stdout
		fmt.Printf("%s\n", formatScan(dInterval, rpmCfg, scan))

	}
	cancel()
	wg.Wait()
	rlog.NoticeMsg("poll exiting")

}

// func init() {
// 	rootCmd.AddCommand(pollCmd)

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// pollCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// pollCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
