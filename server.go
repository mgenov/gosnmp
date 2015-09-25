// Copyright 2012-2015 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

package gosnmp

import (
	"log"
	"net"
	"runtime"
)

const (
	maxTrapSize = 2 << 11 // 2048 bytes
)

// TrapListener defines method that need to be implemented by Trap listeners.
// If OnTRAP panics, the server (caller of OnTRAP) assumes that affect of the panic
// is temporary and recovers by the panic and logs trace to the error log.
type TrapListener interface {
	OnTRAP(trap *SnmpPacket)
}

// A Server defines parameters for running of TRAP daemon that listens for incoming
// trap messages.
type Server struct {
	// Addr is the address in format "localhost:5000" on which server will listen
	Addr string

	// Trap Listener
	Listener TrapListener

	// Error Logger which will be used for logging of default errors
	ErrorLog *log.Logger
}

// Listen listens on UDP network address and then calls Serve with TrapListener to dispatch received TRAP messages to it.
func (s *Server) Listen(address string, listener TrapListener) error {
	s.Addr = address
	addr, err := net.ResolveUDPAddr("udp4", s.Addr)
	if err != nil {
		return err
	}

	s.Listener = listener

	l, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}

	s.Serve(l)

	return nil
}

func (s *Server) Serve(l *net.UDPConn) {
	defer l.Close()

	for {
		buf := make([]byte, maxTrapSize)

		//TODO(mgenov): handle properly read timeout
		n, err := l.Read(buf)

		// TODO(mgenov) handle timeout errors separetely for logging
		if err != nil {
			continue
		}

		// Bigger messages are skipped for processing.
		if n == maxTrapSize {
			continue
		}

		go s.handle(buf[0:n])
	}
}

// handle a newly received trap
func (s *Server) handle(buf []byte) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			logBuf := make([]byte, size)
			logBuf = logBuf[:runtime.Stack(logBuf, false)]
			s.logf("trap: panic while listening %v: %v\n%s", s.Addr, err, logBuf)

		}
	}()

	if s.Listener == nil {
		s.logf("trap: listener is not attached and trap information cannot be dispatched.")
		return
	}

	//Decode received trap and pass it to the listener
	packet := &SnmpPacket{}

	/*
		_, err := unmarshalTrapPDU(buf, packet)
		if err != nil {
			s.logf("trap: unmarshal error %v", err)
		}
	*/

	s.Listener.OnTRAP(packet)
}

func (s *Server) logf(format string, args ...interface{}) {
	if s.ErrorLog != nil {
		s.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
