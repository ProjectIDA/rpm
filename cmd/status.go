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
	"fmt"
	"log"
	"rpm/config"
	rlog "rpm/log"
	"rpm/tycon"
	"strconv"
	"time"
	// "github.com/spf13/cobra"
)

// statusCmd represents the status command
// var statusCmd = &cobra.Command{
// 	Use:   "status",
// 	Short: "A brief description of your command",
// 	Long: `A longer description that spans multiple lines and likely contains examples
// and usage of using your command. For example:

// Cobra is a CLI library for Go that empowers applications.
// This application is a tool to generate the needed files
// to quickly create a Cobra application.`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		runStatusCmd(args)
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(statusCmd)

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }

// Status runs the status command
func Status(host, port string, rpmCfg *config.RPMConfig, pollArgs []string) {

	// rlog.DebugMsg(fmt.Sprintf("status cmd with args[]: %v\n", args))

	// hostport := args[0]
	// host, port := formatSNMPHostPort(hostport)
	cfg.Host = host
	cfg.Port = port
	cfg.RPMCfg = rpmCfg

	rlog.NoticeMsg(fmt.Sprintf("running status command on host: %s:%s\n", cfg.Host, cfg.Port))

	initOids(cfg.RPMCfg)

	tp2din := tycon.NewTPDin2()
	err := tp2din.Initialize(cfg.Host, cfg.Port, 0)
	if err != nil {
		rlog.ErrMsg("unknown error initializing structures for %s:%s, quitting", cfg.Host, cfg.Port)
		log.Fatalln(err)
	}
	err = tp2din.Connect()
	if err != nil {
		rlog.CritMsg("could not connect to %s:%s, quitting", cfg.Host, cfg.Port)
		log.Fatalln(fmt.Errorf("could not connect to %s:%s, quitting", cfg.Host, cfg.Port))
	}
	defer tp2din.SNMPParams.Conn.Close()

	ts, results, err := tp2din.QueryOids(&allOids)
	if err != nil {
		fmt.Printf("error querying device %s:%s\n", cfg.Host, cfg.Port)
		rlog.ErrMsg("error querying device %s:%s", cfg.Host, cfg.Port)
		return
	}

	fmt.Println()
	fmt.Printf("%40s:  %s:%s\n", "Host", cfg.Host, cfg.Port)

	displayStatusInfo(ts, results)

}

func displayStatusInfo(ts time.Time, results map[string]string) {

	for _, val := range cfg.RPMCfg.Oids.Static {
		fmt.Printf("%40s:  %s\n", val.Label, results[val.Oid])
	}
	fmt.Printf("%40s:  %s\n", "Time of Query", ts.Format("2006-01-02 15:04:05 MST"))
	fmt.Println() /// Mon Jan 2 15:04:05 MST 2006

	for _, val := range cfg.RPMCfg.Oids.Relays {
		fmt.Printf("%40s:  %s\n", val.Label, results[val.Oid])
	}
	fmt.Println()

	for _, val := range cfg.RPMCfg.Oids.Voltages {
		volts, _ := strconv.ParseFloat(results[val.Oid], 64)
		fmt.Printf("%40s:  %4.1f (volts)\n", val.Label, volts/10)
	}
	fmt.Println()

	for _, val := range cfg.RPMCfg.Oids.Currents {
		amps, _ := strconv.ParseFloat(results[val.Oid], 64)
		fmt.Printf("%40s:  %4.1f (amps)\n", val.Label, amps/10)
	}
	fmt.Println()

	for _, val := range cfg.RPMCfg.Oids.Temps {
		temp, _ := strconv.ParseFloat(results[val.Oid], 64)
		fmt.Printf("%40s:  %4.1f (deg celsius)\n", val.Label, temp/10)
	}

}
