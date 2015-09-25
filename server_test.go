// Copyright 2012-2014 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

package gosnmp

import (
	"log"
	"os"
	"testing"
)

type nopHandler struct {
}

func (n *nopHandler) OnTRAP(s *SnmpPacket) {
	log.Println("Trap was received.")
}

func TestListen(t *testing.T) {
	slog = log.New(os.Stdout, "", 0)
	LoggingDisabled = true

	nh := &nopHandler{}
	s := &Server{}
	s.Listen(":5000", nh)

}
