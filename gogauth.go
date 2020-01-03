// Package main (gogauth.go) :
// This file is included all commands and options.
package main

import (
	"os"

	"github.com/urfave/cli"
)

// main :
func main() {
	app := cli.NewApp()
	app.Name = "gogauth"
	app.Authors = []*cli.Author{
		{Name: "tanaike [ https://github.com/tanaikech/gogauth ] ", Email: "tanaike@hotmail.com"},
	}
	app.UsageText = "Retrieves accesstoken for using Drive API from Google."
	app.Version = "2.0.2"
	app.Commands = []*cli.Command{
		{
			Name:    "getaccesstoken",
			Aliases: []string{"g"},
			Usage:   "Get accesstoken",
			Action:  getTokens,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "reauth, r",
					Aliases: []string{"r"},
					Usage:   "Retrieve access and refresh tokens. If you changed scopes, please use this.",
				},
				&cli.IntFlag{
					Name:    "port, p",
					Aliases: []string{"p"},
					Usage:   "Port number of temporal web server for retrieving authorization code.",
					Value:   8080,
				},
			},
		},
		{
			Name:    "checkaccesstoken",
			Aliases: []string{"c"},
			Usage:   "Check accesstoken",
			Action:  checkAccesstoken,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "accesstoken, a",
					Aliases: []string{"a"},
					Usage:   "Confirm condition of current accesstoken",
				},
			},
		},
	}
	app.CommandNotFound = commandNotFound
	app.Run(os.Args)
}
