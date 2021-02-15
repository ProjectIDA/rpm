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
	rlog "rpm/log"
	"rpm/tycon"
	"time"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStatusCmd(args)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runStatusCmd(args []string) {

	rlog.DebugMsg(fmt.Sprintf("status cmd with args[]: %v\n", args))

	hostport := args[0]

	rlog.NoticeMsg(fmt.Sprintf("host: %s\n", hostport))

	initOids(rpmCfg)

	tp2din := tycon.NewTPDin2()
	err := tp2din.Initialize(hostport, 0, allOids)
	if err != nil {
		rlog.ErrMsg("unknown error initializing tp2din... quitting")
		log.Fatalln(err)
	}
	err = tp2din.Connect()
	if err != nil {
		rlog.CritMsg("could not connect to %s... quitting.", hostport)
		log.Fatalln(fmt.Errorf("could not connect to %s... quitting", hostport))
	}
	defer tp2din.SNMPParams.Conn.Close()

	ts, results, err := tp2din.QueryOids()
	if err != nil {
		rlog.ErrMsg("error querying device")
	}

	fmt.Printf("\n%s\n\n", ts.Format(time.RFC1123))

	// for _, group := range [][]config.OidInfo{
	// 	rpmCfg.Oids.Static,
	// 	rpmCfg.Oids.Relays,
	// 	rpmCfg.Oids.Voltages,
	// 	rpmCfg.Oids.Currents,
	// 	rpmCfg.Oids.Temps,
	// } {
	// 	for _, val := range group {
	// 		fmt.Printf("%40s:  %4s\n", val.Label, results[val.Oid])
	// 	}
	// 	fmt.Println()
	// }

	for _, val := range rpmCfg.Oids.Static {
		fmt.Printf("%40s:  %4s\n", val.Label, results[val.Oid])
	}
	fmt.Println()

	for _, val := range rpmCfg.Oids.Relays {
		fmt.Printf("%40s:  %4s\n", val.Label, results[val.Oid])
	}
	fmt.Println()

	for _, val := range rpmCfg.Oids.Voltages {
		fmt.Printf("%40s:  %4s (volts x 10)\n", val.Label, results[val.Oid])
	}
	fmt.Println()

	for _, val := range rpmCfg.Oids.Currents {
		fmt.Printf("%40s:  %4s (amps x 10)\n", val.Label, results[val.Oid])
	}
	fmt.Println()

	for _, val := range rpmCfg.Oids.Temps {
		fmt.Printf("%40s:  %4s (deg celsius x 10)\n", val.Label, results[val.Oid])
	}

}
