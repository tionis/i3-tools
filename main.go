// Copyright 2018 Google Inc.
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

// i3status is a port of the default i3status configuration to barista.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"go.i3wm.org/i3/v4"
	"log"
	"math/rand"
	"os"
)

func main() {
	app := &cli.App{
		Name: "i3-tools",
		Commands: []*cli.Command{
			{
				Name:  "bar",
				Usage: "bar commands",
				Subcommands: []*cli.Command{
					{
						Name:  "render",
						Usage: "render bar output as json for i3bar",
						Action: func(c *cli.Context) error {
							return i3status(defaultBar())
						},
					},
				},
			},
			{
				Name:  "api",
				Usage: "access to the i3 api",
				Subcommands: []*cli.Command{
					{
						Name: "GetWorkspaces",
						Usage: "GetWorkspaces returns i3’s current workspaces.\n" +
							"GetWorkspaces is supported in i3 ≥ v4.0 (2011-07-31).",
						Action: func(c *cli.Context) error {
							workspaces, err := i3.GetWorkspaces()
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(workspaces, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "GetTree",
						Usage: "GetTree returns i3’s layout tree.\n" +
							"GetTree is supported in i3 ≥ v4.0 (2011-07-31).",
						Action: func(c *cli.Context) error {
							tree, err := i3.GetTree()
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(tree, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "GetVersion",
						Usage: "GetVersion returns i3’s version.\n" +
							"GetVersion is supported in i3 ≥ v4.3 (2012-09-19).",
						Action: func(c *cli.Context) error {
							version, err := i3.GetVersion()
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(version, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "GetOutputs",
						Usage: "GetOutputs returns i3’s current outputs.\n" +
							"GetOutputs is supported in i3 ≥ v4.0 (2011-07-31).",
						Action: func(c *cli.Context) error {
							outputs, err := i3.GetOutputs()
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(outputs, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "GetBarIDs",
						Usage: "GetBarIDs returns an array of configured bar IDs.\n" +
							"GetBarIDs is supported in i3 ≥ v4.1 (2011-11-11).",
						Action: func(c *cli.Context) error {
							ids, err := i3.GetBarIDs()
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(ids, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "GetBarConfig",
						Usage: "GetBarConfig returns the configuration for the " +
							"bar with the specified barID.\n" +
							"Obtain the barID from GetBarIDs.\n" +
							"GetBarConfig is supported in i3 ≥ v4.1 (2011-11-11).",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Usage:    "id of the bar to get the config for",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							config, err := i3.GetBarConfig(c.String("id"))
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(config, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "GetBindingModes",
						Usage: "GetBindingModes returns the names of all currently configured binding modes.\n" +
							"GetBindingModes is supported in i3 ≥ v4.13 (2016-11-08).",
						Action: func(c *cli.Context) error {
							modes, err := i3.GetBindingModes()
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(modes, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "SendTick",
						Usage: "SendTick sends a tick event with the provided payload.\n" +
							"SendTick is supported in i3 ≥ v4.15 (2018-03-10).",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "payload",
								Usage:    "payload to send with the tick",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							result, err := i3.SendTick(c.String("payload"))
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(result, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "Sync",
						Usage: "Sync sends a tick event with the provided payload.\n" +
							"Sync is supported in i3 ≥ v4.16 (2018-11-04).",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:     "window",
								Usage:    "window to sync",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							syncRequest := i3.SyncRequest{
								Window: uint32(c.Int("window")),
								Rnd:    rand.Uint32(),
							}
							result, err := i3.Sync(syncRequest)
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(result, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "GetMarks",
						Usage: "GetMarks returns the names of all currently set marks.\n" +
							"GetMarks is supported in i3 ≥ v4.1 (2011-11-11).",
						Action: func(c *cli.Context) error {
							marks, err := i3.GetMarks()
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(marks, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "GetBindingState",
						Usage: "GetBindingState returns the currently active binding mode.\n" +
							"GetBindingState is supported in i3 ≥ 4.19 (2020-11-15)",
						Action: func(c *cli.Context) error {
							state, err := i3.GetBindingState()
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(state, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "Restart",
						Usage: "Restart sends the restart command to i3. " +
							"Sending restart via RunCommand will result in " +
							"a deadlock: since i3 restarts before it sends the " +
							"reply to the restart command, RunCommand will retry " +
							"the command indefinitely.\n" +
							"Restart is supported in i3 ≥ v4.14 (2017-09-04).",
						Action: func(c *cli.Context) error {
							return i3.Restart()
						},
					},
					{
						Name: "RunCommand",
						Usage: "RunCommand makes i3 run the specified command.\n" +
							"Error is non-nil if any CommandResult.Success is not true. " +
							"See IsUnsuccessful if you send commands which are expected to " +
							"fail.\nRunCommand is supported in i3 ≥ v4.0 (2011-07-31).",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "command",
								Usage:    "command to send to i3",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							result, err := i3.RunCommand(c.String("command"))
							if err != nil {
								return err
							}
							marshalled, err := json.MarshalIndent(result, "", "  ")
							if err != nil {
								return err
							}
							fmt.Println(string(marshalled))
							return nil
						},
					},
					{
						Name: "Subscribe",
						Usage: "Subscribe returns an EventReceiver for receiving " +
							"events of the specified types from i3.\n" +
							"Unless the ordering of events matters to your use-case, " +
							"you are encouraged to call Subscribe once per event type, " +
							"so that you can use type assertions instead of type switches.\n" +
							"Subscribe is supported in i3 ≥ v4.0 (2011-07-31).",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "mode",
								Usage: "subscribe to mode events",
							},
							&cli.BoolFlag{
								Name:  "barconfig-update",
								Usage: "subscribe to barconfig update events",
							},
							&cli.BoolFlag{
								Name:  "binding",
								Usage: "subscribe to binding events",
							},
							&cli.BoolFlag{
								Name:  "tick",
								Usage: "subscribe to tick events",
							},
							&cli.BoolFlag{
								Name:  "workspace",
								Usage: "subscribe to workspace events",
							},
							&cli.BoolFlag{
								Name:  "output",
								Usage: "subscribe to output events",
							},
							&cli.BoolFlag{
								Name:  "window",
								Usage: "subscribe to window events",
							},
							&cli.BoolFlag{
								Name:  "shutdown",
								Usage: "subscribe to shutdown events",
							},
						},
						Action: func(context *cli.Context) error {
							eventTypes := make([]i3.EventType, 0)
							if context.Bool("mode") {
								eventTypes = append(eventTypes, i3.ModeEventType)
							}
							if context.Bool("barconfig-update") {
								eventTypes = append(eventTypes, i3.BarconfigUpdateEventType)
							}
							if context.Bool("binding") {
								eventTypes = append(eventTypes, i3.BindingEventType)
							}
							if context.Bool("tick") {
								eventTypes = append(eventTypes, i3.TickEventType)
							}
							if context.Bool("workspace") {
								eventTypes = append(eventTypes, i3.WorkspaceEventType)
							}
							if context.Bool("output") {
								eventTypes = append(eventTypes, i3.OutputEventType)
							}
							if context.Bool("windows") {
								eventTypes = append(eventTypes, i3.WindowEventType)
							}
							if context.Bool("shutdown") {
								eventTypes = append(eventTypes, i3.ShutdownEventType)
							}
							receiver := i3.Subscribe(eventTypes...)
							for receiver.Next() {
								marshalled, err := json.Marshal(receiver.Event())
								if err != nil {
									return err
								}
								fmt.Println(string(marshalled))
							}
							return receiver.Close()
						},
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to run app: %v", errors.Unwrap(err))
	}
}
