// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

var (
	version  = "0.1.0"
	addr     = flag.String("addr", "0.0.0.0:9876", "server address")
	vacation = flag.String("vacation", "", "start vacation until the specified date")
)

func init() {
	log.SetPrefix("office: ")
	log.SetFlags(log.Lmsgprefix | log.LstdFlags | log.Lshortfile)
	flag.Usage = usage
	flag.Parse()
}

func usage() {
	fmt.Fprintf(os.Stderr,
		`office is a command that exposes Changkun's office status to the
public. The status can be fetched via: https://changkun.de/x/working?

Version: %s
GoVersion: %s

Command line usage:

$ office [-vacation <time>]

options:

office -vacation
	Vacation mode

examples:

office
	$ curl -L https://changkun.de/s/working
	Yes!
	$ curl -L https://changkun.de/s/working
	No, he left 10s ago.

office -vacation 2021-08-11

	$ curl -L https://changkun.de/s/working
	No, he is on vacation and will return on 11 Aug.
`, version, runtime.Version())
	os.Exit(2)
}

func main() {
	if *vacation != "" {
		t, err := time.Parse("2006-01-02", *vacation)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid time parameter")
			flag.Usage()
			return
		}
		status.StartVacation(t)
	}

	serve()
}
