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
	"errors"
	"github.com/urfave/cli/v2"
	"log"
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
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to run app: %v", errors.Unwrap(err))
	}
}
