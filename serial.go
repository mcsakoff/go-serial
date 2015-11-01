// Copyright (c) 2015, Nick Patavalis (npat@efault.net).
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can
// be found in the LICENSE file.

// Package serial provides a simple system-independent interface for
// accessing serial ports.
package serial

import "time"

// Port is a serial port
type Port struct {
	Name string // Name used at Port.Open
	*port
}

// ParityMode encodes the supported bit-parity modes
type ParityMode int

const (
	ParityNone  ParityMode = iota // No parity bit
	ParityEven                    // Even bit-parity
	ParityOdd                     // Odd bit-parity
	ParityMark                    // Parity bit to logical 1 (mark)
	ParitySpace                   // Parity bit to logical 0 (space)
)

// FlowMode encodes the supported flow-control modes
type FlowMode int

const (
	FlowNone    FlowMode = iota // No flow control
	FlowRTSCTS                  // Hardware flow control
	FlowXONXOFF                 // Software flow control
	FlowOther                   // Unknown mode
)

//go:generate stringer -type=ParityMode,FlowMode -output enum_strings.go

// Conf is used to pass the serial port's configuration parameters to
// and from methods of this package.
type Conf struct {
	Baudrate int        // in Bits Per Second
	Databits int        // 5, 6, 7, or 8
	Stopbits int        // 1 or 2
	Parity   ParityMode // see ParityXXX constants
	Flow     FlowMode   // see FlowXXX constants
	NoReset  bool       // don't reset and don't hangup on close
}

// Functions bellow are actually implemented in the system-specific
// serial_<system>.go files.

// Open opens the named serial port. Open records the port-settings
// (so they can be reset at Close), and sets the port to what unix
// calls "raw-mode" (transparent operation, without character
// translation or other processing). Other port settings (baudratre,
// character format, flow-control, etc.) are not altered.
func Open(name string) (port *Port, err error) {
	p, err := open(name)
	if err != nil {
		return nil, err
	}
	return &Port{Name: name, port: p}, nil
}

// Close closes the port. Unless the port has been configured with
// Conf.NoReset = true, the port is reset to its original settings
// (the ones it had at open), and the connection is terminated by
// de-asserting the modem control lines.
func (p *Port) Close() error {
	return p.port.close()
}

// GetConf returns the serial port's configuration parameters as a
// Conf structure.
func (p *Port) GetConf() (conf Conf, err error) {
	return p.port.getConf()
}

// ConfFlags are flags controlling which parameters to configure
type ConfFlags int

const (
	ConfBaudrate ConfFlags = 1 << iota
	ConfDatabits
	ConfParity
	ConfStopbits
	ConfFlow
	ConfNoReset
	ConfFormat = ConfDatabits | ConfParity | ConfStopbits
	ConfAll    = ConfBaudrate | ConfFormat | ConfFlow | ConfNoReset
)

// ConfSome configures the serial port using some of the parameters in
// the Conf structure, based on the value of the flags argument.
func (p *Port) ConfSome(conf Conf, flags ConfFlags) error {
	return p.port.confSome(conf, flags)
}

// Conf configures the serial port using the parameters in the Conf
// structure
func (p *Port) Conf(conf Conf) error {
	return p.port.confSome(conf, ConfAll)
}

// Read is compatible with the Read method of the io.Reader
// interface. In addition Read honors the timeout set by
// Port.SetDeadline and Port.SetReadDeadline. If no data are read
// before the timeout expires Read returns with err == ErrTimeout (and
// n == 0).
func (p *Port) Read(b []byte) (n int, err error) {
	return p.port.read(b)
}

// Write is compatible with the Write method of the io.Writer
// interface. In addition Write honors the timeout set by
// Port.SetDeadline and Port.SetWriteDeadline. If less than len(p)
// data are writen before the timeout expires Write returns with err
// == ErrTimeout (and n < len(p)).
func (p *Port) Write(b []byte) (n int, err error) {
	return p.port.write(b)
}

// SetDeadline sets the deadline for both Read and Write operations on
// the port. Deadlines are expressed as ABSOLUTE instances in
// time. For example, to set a deadline 5 seconds to the future do:
//
//   p.SetDeadline(time.Now().Add(5 * time.Second))
//
// This is equivalent to:
//
//   p.SetReadDeadline(time.Now().Add(5 * time.Second))
//   p.SetWriteDeadline(time.Now().Add(5 * time.Second))
//
// A zero value for t, cancels (removes) the existing deadline.
//
func (p *Port) SetDeadline(t time.Time) error {
	return p.port.setDeadline(t)
}

// SetReadDeadline sets the deadline for Read operations. See also
// SetDeadline.
func (p *Port) SetReadDeadline(t time.Time) error {
	return p.port.setReadDeadline(t)
}

// SetWriteDeadline sets the deadline for Write operations. See also
// SetDeadline.
func (p *Port) SetWriteDeadline(t time.Time) error {
	return p.port.setWriteDeadline(t)
}

type flushSel int

const (
	flushIn flushSel = iota
	flushOut
	flushInOut
)

func (p *Port) Flush() error {
	return p.port.flush(flushInOut)
}

func (p *Port) FlushIn() error {
	return p.port.flush(flushIn)
}

func (p *Port) FlushOut() error {
	return p.port.flush(flushOut)
}

// speedTable is used to map numeric tty speeds (baudrates) to the
// respective code (Bxxx) values.
type speedTable []struct {
	speed int
	code  uint32
}

func (t speedTable) Code(speed int) (code uint32, ok bool) {
	for _, s := range t {
		if s.speed == speed {
			return s.code, true
		}
	}
	return 0, false
}

func (t speedTable) Speed(code uint32) (speed int, ok bool) {
	for _, s := range t {
		if s.code == code {
			return s.speed, true
		}
	}
	return 0, false
}
