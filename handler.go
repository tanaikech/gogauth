// Package main (handler.go) :
// Handler for gogauth
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
)

// getTokens : Get access token and refresh token
func getTokens(c *cli.Context) error {
	if c.Bool("reauth") {
		defAuthContainer(c).gogauthIni(c).reAuth()
		fmt.Print("Done.")
	} else {
		defAuthContainer(c).gogauthIni(c).goauth().dispResult(c)
	}
	return nil
}

// checkAccesstoken : Check access token
func checkAccesstoken(c *cli.Context) error {
	if len(c.String("accesstoken")) > 0 {
		a := defAuthContainer(c)
		a.gogauthCfg.Accesstoken = c.String("accesstoken")
		exp := a.chkAtoken()
		a.chkAt.Scopenumber = len(strings.Split(a.chkAt.Scope, " "))
		a.chkAt.Expdatetime = fmt.Sprintf("%s", time.Unix(exp, 0))
		dispRes, _ := json.MarshalIndent(a.chkAt, "", "  ")
		fmt.Printf("%+v\n", string(dispRes))
	}
	return nil
}

// dispResult : Display result
func (a *AuthContainer) dispResult(c *cli.Context) {
	fmt.Printf("%s", a.gogauthCfg.Accesstoken)
}

// commandNotFound :
func commandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(os.Stderr, "'%s' is not a %s command. Check '%s --help' or '%s -h'.", command, c.App.Name, c.App.Name, c.App.Name)
	os.Exit(2)
}
