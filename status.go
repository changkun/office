// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"sync/atomic"
	"time"
)

var status MyStatus

type officeStatus = int32

const (
	statusOn officeStatus = iota
	statusOff
	statusVacation
)

// MyStatus represents my work status.
type MyStatus struct {
	status        officeStatus
	lastAvailable atomic.Value
	returnDate    time.Time
}

func (s *MyStatus) IsWorking() (bool, time.Time) {
	switch atomic.LoadInt32(&s.status) {
	case statusOn:
		return true, time.Now()
	case statusOff:
		return false, s.lastAvailable.Load().(time.Time)
	case statusVacation:
		return false, s.returnDate
	}
	panic("invalid status")
}

func (s *MyStatus) IsInMeting() bool {
	in, err := isInMeeting()
	if err == nil && in {
		return true
	}
	return false
}

func (s *MyStatus) StartWork() {
	atomic.CompareAndSwapInt32(&s.status, statusOff, statusOn)
}

func (s *MyStatus) StopWork() {
	s.lastAvailable.Store(time.Now())
	atomic.CompareAndSwapInt32(&s.status, statusOn, statusOff)
}

func (s *MyStatus) StartVacation(until time.Time) {
	s.status = statusVacation
	s.returnDate = until
}

func (s *MyStatus) OnVacation() (bool, time.Time) {
	if atomic.LoadInt32(&s.status) == statusVacation {
		return true, s.returnDate
	}
	return false, time.Now()
}
