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
package main

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
    "os/user"
    "path/filepath"
    "strconv"
    "syscall"
	"rpm/cmd"
	rlog "rpm/log"
)

type ctxKeyType string

func main() {

    var err error

	err = rlog.InitLogging("rpm", syslog.LOG_LOCAL0|syslog.LOG_NOTICE)
	if err != nil {
		fmt.Fprintln(os.Stderr, "rpm: error creating logger (fprintf)")
		log.Fatal("rpm: error creating logger (log.fatal)")
	}

	abspath, err := filepath.Abs(os.Args[0])
	if err != nil {
		rlog.ErrMsg(err.Error())
		os.Exit(1)
	}
	rlog.NoticeMsg(fmt.Sprintf("%s starting up...", abspath))


    var uid, gid int
    nrtsuser, err := user.Lookup("nrts")
    if uid, err = strconv.Atoi(nrtsuser.Uid); err != nil {
        log.Fatal("ERROR converting UID to Int")
    }
    if gid, err = strconv.Atoi(nrtsuser.Gid); err != nil {
        log.Fatal("ERROR converting GID to Int")
    }
    syscall.Setuid(uid)
    syscall.Setgid(gid)
    os.Chdir(nrtsuser.HomeDir)

    var wd string
    wd, err = os.Getwd()
    rlog.NoticeMsg(fmt.Sprintf("working in: %s", wd))

	cmd.Execute()
}
