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

// serve serves the status server.
func serve() {
	updateStatus()

	http.Handle("/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		log.Printf("access from %s\n", ip(r))

		on, _ := status.OnVacation()
		if on {
			checkVacation(w, r)
			return
		}

		checkWork(w, r)
	}))

	log.Printf("running on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Printf("listen and serve failed at %s: %v", *addr, err)
	}
}

// updateStatus does status check if not on vacation.
func updateStatus() {
	on, _ := status.OnVacation()
	if on {
		return
	}

	// initial status
	atomic.StoreInt32(&status.status, statusOff)
	status.lastAvailable.Store(time.Now())
	lock, err := isScreenLocked()
	if err == nil && !lock {
		status.StartWork()
	} else {
		status.StopWork()
	}

	go func() {
		t := time.Tick(10 * time.Second)
		for range t {
			lock, err := isScreenLocked()
			if err == nil && !lock {
				log.Println("he is working")
				status.StartWork()
				continue
			}
			if err != nil {
				log.Printf("lockscreen check err: %s\n", err)
			}
			status.StopWork()
			log.Println("he left the office: ", status.lastAvailable.Load().(time.Time))
		}
	}()
}

// checkWork serves the work status.
func checkWork(w http.ResponseWriter, r *http.Request) {
	working, last := status.IsWorking()
	if !working {
		w.Write([]byte(fmt.Sprintf("No, he left %s ago.", time.Since(last).Round(time.Second))))
		return
	}

	if status.IsInMeting() {
		w.Write([]byte("Yes! But current in a meeting."))
		return
	}

	w.Write([]byte("Yes!"))
}

// checkVacation serves the vacation status.
func checkVacation(w http.ResponseWriter, r *http.Request) {
	var msg string
	_, ret := status.OnVacation()
	if ret.Year() == time.Now().Year() { // same year
		msg = ret.Format("02 Jan")
	} else {
		msg = ret.Format("Jan 2, 2006")
	}

	w.Write([]byte(fmt.Sprintf("No, he is on vacation and will return on %s.", msg)))
}

// isScreenLocked checks if the screensaver is active or not.
//
// Active screensaver means the office computer is not in-use, and
// representing I am not in the office.
func isScreenLocked() (bool, error) {
	// Do this command:
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

// isInMeeting checks if the connected camera is on or not.
//
// If the connected camera is on it basically means I am in a meeting,
// such as zoom. The principle is, a python script tries to open
// the camera, if the camera is acquired by other application, it will
// print out a False and representing I am in the meeting.
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
	return strings.Contains(out.String(), "False"), nil
}

// ip parses the request and returns the source address.
func ip(r *http.Request) string {
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
