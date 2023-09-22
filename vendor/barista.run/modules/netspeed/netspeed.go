// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package netspeed provides an i3bar module to display network utilisation.
package netspeed // import "barista.run/modules/netspeed"

import (
	"time"

	"barista.run/bar"
	"barista.run/base/value"
	"barista.run/format"
	l "barista.run/logging"
	"barista.run/outputs"
	"barista.run/timing"

	"github.com/martinlindhe/unit"
	"github.com/vishvananda/netlink"
)

// Speeds represents bidirectional network traffic.
type Speeds struct {
	Rx, Tx unit.Datarate
	// Keep track of whether these speeds are actually 0
	// or uninitialised.
	available bool
	state     netlink.LinkOperState
}

// Total gets the total speed (both up and down).
func (s Speeds) Total() unit.Datarate {
	return s.Rx + s.Tx
}

// Connected returns true if the network is connected.
func (s Speeds) Connected() bool {
	return s.state >= 5 // IF_OPER_DORMANT
}

// Module represents a netspeed bar module. It supports setting the output
// format, click handler, and update frequency.
type Module struct {
	iface      string
	scheduler  *timing.Scheduler
	outputFunc value.Value // of func(Speeds) bar.Output
}

// New constructs an instance of the netspeed module for the given interface.
func New(iface string) *Module {
	m := &Module{
		iface:     iface,
		scheduler: timing.NewScheduler(),
	}
	l.Label(m, iface)
	l.Register(m, "scheduler", "outputFunc")
	m.RefreshInterval(3 * time.Second)
	// Default output is just the up and down speeds in SI.
	m.Output(func(s Speeds) bar.Output {
		return outputs.Textf("%s up | %s down",
			format.IByterate(s.Tx), format.IByterate(s.Rx))
	})
	return m
}

// Output configures a module to display the output of a user-defined function.
func (m *Module) Output(outputFunc func(Speeds) bar.Output) *Module {
	m.outputFunc.Set(outputFunc)
	return m
}

// RefreshInterval configures the polling frequency for network speed.
// Since there is no concept of an instantaneous network speed, the speeds will
// be averaged over this interval before being displayed.
func (m *Module) RefreshInterval(interval time.Duration) *Module {
	m.scheduler.Every(interval)
	return m
}

// For tests.
var linkByName = netlink.LinkByName

// Stream starts the module.
func (m *Module) Stream(s bar.Sink) {
	lastRead := timing.Now()
	lastRx, lastTx, _, err := linkRxTxState(m.iface)
	if s.Error(err) {
		return
	}

	var speeds Speeds
	outputFunc := m.outputFunc.Get().(func(Speeds) bar.Output)
	nextOutputFunc, done := m.outputFunc.Subscribe()
	defer done()

	for {
		if speeds.available {
			s.Output(outputFunc(speeds))
		}
		select {
		case <-nextOutputFunc:
			outputFunc = m.outputFunc.Get().(func(Speeds) bar.Output)
		case <-m.scheduler.C:
			rx, tx, state, err := linkRxTxState(m.iface)
			if s.Error(err) {
				return
			}
			now := timing.Now()
			duration := now.Sub(lastRead).Seconds()

			speeds.available = true
			speeds.Rx = unit.Datarate(float64(rx-lastRx)/duration) * unit.BytePerSecond
			speeds.Tx = unit.Datarate(float64(tx-lastTx)/duration) * unit.BytePerSecond
			speeds.state = state

			lastRead = now
			lastRx = rx
			lastTx = tx
		}
	}
}

func linkRxTxState(iface string) (rx, tx uint64, state netlink.LinkOperState, err error) {
	var link netlink.Link
	link, err = linkByName(iface)
	if err != nil {
		return
	}
	state = link.Attrs().OperState
	linkStats := link.Attrs().Statistics
	rx = linkStats.RxBytes
	tx = linkStats.TxBytes
	return
}
