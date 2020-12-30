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
	"os"
	"os/signal"
	"rpm/tycon"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// pollCmd represents the poll command
var pollCmd = &cobra.Command{
	Use:   "poll host_and_port polling_interval_in_secs",
	Short: "Poll SNMP target for values",
	Long:  `Poll SNMP target for values at a fixed interval >= 1.0 seconds`,
	Args:  checkArgs, //cobra.ExactArgs(2),
	Run:   poll,
}

func checkArgs(cmd *cobra.Command, args []string) error {

	numArgs := 2
	if len(args) != numArgs {
		return fmt.Errorf("poll requires %d arguments", numArgs)
	}
	if _, err := strconv.ParseFloat(args[1], 32); err != nil {
		return fmt.Errorf("Invalid polling interval, must be a positive number >= 0.5: %s", args[1])
	}

	return nil
}

// SetupSignals to trap for external kill signals
func SetupSignals(sigs ...os.Signal) chan bool {

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

// func exportScan(scan *tycon.TPDin2ScanData) {

// 	fmt.Printf("%v:: ", scan.TS)
// 	for oid, val := range scan.Data {
// 		fmt.Printf("%s == %s\n", oid, val)
// 	}
// 	fmt.Println()

// }

func poll(cmd *cobra.Command, args []string) {
	// snmpwalk -On -c readwrite -M /usr/local/share/snmp/mibs -v 1 localhost

	fmt.Printf("poll called with args[]: %v\n", args)
	// fmt.Println("let's make a tpdin2 object...")

	hostport := args[0]
	fInterval, _ := strconv.ParseFloat(args[1], 32)

	fmt.Printf("Host: %s; interval: %f\n", hostport, fInterval)

	tp2din := tycon.NewTPDin2()

	dInterval := time.Duration(fInterval) * time.Second
	hInterval := dInterval / 2
	err := tp2din.Initialize(hostport, dInterval, rpmCfg)
	if err != nil {
		log.Fatalln(err)
	}

	err = tp2din.Connect()
	if err != nil {
		log.Fatal(fmt.Errorf("could not connect to %s", hostport))
	}

	sigdone := SetupSignals(syscall.SIGINT, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	err = tp2din.PollStart(ctx, &wg)
	if err != nil {
		cancel()
		wg.Wait()
		log.Fatal(err)
	}

	var scan *tycon.TPDin2Scan
	targetTime := time.Now().Round(1 * time.Second).Add(3 * time.Second)

	var offset time.Duration
	first := true
	exiting := false
	for !exiting {

		targetTime = targetTime.Add(dInterval)
		fmt.Printf("target time: %v\n", targetTime.String())

		// In case loop iteration took more the 1 time interval
		for targetTime.Before(time.Now()) {
			targetTime = targetTime.Add(dInterval)
		}

		select {

		case <-time.After(time.Until(targetTime)):

			scan = tp2din.GetScan()
			if scan == nil {
				fmt.Printf("No scan available\n")
				first = true
				continue
			}

		case <-sigdone:
			fmt.Println("got signal")
			exiting = true
			continue
		}

		if first {
			fmt.Println("initial scan received")
			first = false
		}

		offset = scan.TS.Sub(targetTime)

		// see if scan should go with next second, perhaps due to network lag
		if (offset > hInterval) && (offset < (dInterval + hInterval)) {
			fmt.Println("warning: delayed scan. Skipping one interval")
			targetTime = targetTime.Add(dInterval)
		} else if offset < -hInterval {
			fmt.Printf("error: unexpected timestamp: %v; target time: %v; dropping scan\n", scan.TS, targetTime)
			first = true
			continue
		}

		ts := scan.TS
		data := scan.Data
		fmt.Printf("Target time: %s\n", targetTime.String())
		fmt.Printf("Scan time:   %s\n", ts.String())
		fmt.Printf("Data: %v\n", data)
	}
	cancel()
	wg.Wait()

}

func init() {
	rootCmd.AddCommand(pollCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pollCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pollCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// TargetDevice returns the target device to poll in host:port format
func TargetDevice() (host string, err error) {

	return "", nil
}
