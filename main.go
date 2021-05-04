// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

func isInMeeting() (bool, error) {
	// python check.py | grep "False"
	cmd := exec.Command("python", "check.py")
	var (
		out    bytes.Buffer
		outErr bytes.Buffer
	)
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("%w: %v", err, outErr.String())
	}
	if !strings.Contains(out.String(), "False") {
		return false, nil
	}
	return true, nil
}

func isScreenLocked() (bool, error) {
	// gnome-screensaver-command -q | grep "is active"
	cmd := exec.Command("gnome-screensaver-command", "-q")
	var (
		out    bytes.Buffer
		outErr bytes.Buffer
	)
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("%w: %v", err, outErr.String())
	}
	if !strings.Contains(out.String(), "is active") {
		return false, nil
	}
	return true, nil
}

func addr(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	if ip == "" {
		ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	}
	if ip != "" {
		return ip
	}
	ip = r.Header.Get("X-Appengine-Remote-Addr")
	if ip != "" {
		return ip
	}
	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return "unknown" // use unknown to guarantee non empty string
	}
	return ip
}

var (
	inOffice      int32
	lastAvailable atomic.Value // time.Time
)

func main() {
	log.SetPrefix("office: ")
	log.SetFlags(log.Lmsgprefix | log.LstdFlags | log.Lshortfile)

	http.Handle("/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		log.Printf("access from %s\n", addr(r))

		v := atomic.LoadInt32(&inOffice)
		if v == 1 {
			t := lastAvailable.Load().(time.Time)
			w.Write([]byte(fmt.Sprintf("No, he had already left for %s.", time.Since(t).Round(time.Second))))
			return
		}

		in, err := isInMeeting()
		if err == nil && in {
			w.Write([]byte("Yes! But current in a meeting."))
			return
		}
		w.Write([]byte("Yes!"))
	}))

	go func() {
		lastAvailable.Store(time.Now())
		t := time.Tick(10 * time.Second)
		for range t {

			lock, err := isScreenLocked()
			if err == nil {
				if !lock {
					lastAvailable.Store(time.Now())
					atomic.SwapInt32(&inOffice, 0) // in office
					continue
				}
			}
			if err != nil {
				fmt.Printf("lockscreen check err: %s\n", err)
			}
			atomic.SwapInt32(&inOffice, 1) // not in office
		}
	}()

	const addr = "0.0.0.0:9876"
	log.Println("running")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Listen and serve at %s: %v", addr, err)
	}
}
