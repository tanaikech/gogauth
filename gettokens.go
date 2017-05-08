// Package main (gettokens.go) :
// Get access token and refresh token from client_secret.json and scopes.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tanaikech/gogauth/utl"
	"github.com/urfave/cli"
)

// const :
const (
	appname          = "gogauth"
	clientsecretFile = "client_secret.json"
	cfgFile          = "gogauth.cfg"
	oauthurl         = "https://accounts.google.com/o/oauth2/"
	chkatutl         = "https://www.googleapis.com/oauth2/v3/"
	initscope        = "https://www.googleapis.com/auth/drive.readonly"
)

// InitVal : Initial values
type InitVal struct {
	pstart  time.Time
	workdir string
	Port    int
	update  bool
}

// gogauthCfg : Configuration file for gogauth
type gogauthCfg struct {
	Clientid     string   `json:"client_id"`
	Clientsecret string   `json:"client_secret"`
	Refreshtoken string   `json:"refresh_token"`
	Accesstoken  string   `json:"access_token,omitempty"`
	Expiresin    int64    `json:"expires_in,omitempty"`
	Scopes       []string `json:"scopes"`
}

// Cinstalled : File of client-secret.json
type Cinstalled struct {
	ClientID                string   `json:"client_id"`
	Projectid               string   `json:"project_id"`
	Authuri                 string   `json:"auth_uri"`
	Tokenuri                string   `json:"token_uri"`
	Authproviderx509certurl string   `json:"auth_provider_x509_cert_url"`
	Clientsecret            string   `json:"client_secret"`
	Redirecturis            []string `json:"redirect_uris"`
}

// Cs : Client_secret.json
type Cs struct {
	Cid Cinstalled `json:"installed,omitempty"`
	Ciw Cinstalled `json:"web,omitempty"`
}

// Atoken : Accesstoken given from Google
type Atoken struct {
	Accesstoken  string `json:"access_token"`
	Refreshtoken string `json:"refresh_token"`
	Expiresin    int64  `json:"expires_in"`
}

// chkAt : Information of accesstoken
type chkAt struct {
	Azu           string `json:"azu,omitempty"`
	Aud           string `json:"aud,omitempty"`
	Scope         string `json:"scope,omitempty"`
	Scopenumber   int    `json:"scope_number,omitempty"`
	Exp           string `json:"exp,omitempty"`
	Expdatetime   string `json:"exp_datetime,omitempty"`
	Expiresin     string `json:"expires_in,omitempty"`
	Email         string `json:"email,omitempty"`
	Emailverified string `json:"email_verified,omitempty"`
	Accesstype    string `json:"access_type,omitempty"`
	Error         string `json:"error_description,omitempty"`
}

// serverInfToGetCode : For getting auth code
type serverInfToGetCode struct {
	Response chan authCode
	Start    chan bool
	End      chan bool
}

// authCode : For getting auth code
type authCode struct {
	Code string
	Err  error
}

// AuthContainer : Struct container for using OAuth2
type AuthContainer struct {
	*InitVal    // Initial values
	*gogauthCfg // Config for gogauth
	*Cs         // Client_secret.json
	*Atoken     // Accesstoken from Google
	*chkAt      // Check accesstoken
}

// DefAuthContainer : Struct container for authorization
func defAuthContainer(c *cli.Context) *AuthContainer {
	var err error
	a := &AuthContainer{
		&InitVal{},
		&gogauthCfg{},
		&Cs{},
		&Atoken{},
		&chkAt{},
	}
	a.InitVal.pstart = time.Now()
	a.InitVal.workdir, err = filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	a.InitVal.Port = c.Int("port")
	a.gogauthCfg.Scopes = []string{initscope}
	return a
}

// gogauthIni : Initialize
func (a *AuthContainer) gogauthIni(c *cli.Context) *AuthContainer {
	if cfgdata, err := ioutil.ReadFile(filepath.Join(a.InitVal.workdir, cfgFile)); err == nil {
		err = json.Unmarshal(cfgdata, &a.gogauthCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Format error of '%s'. ", cfgFile)
			os.Exit(1)
		}
	} else {
		return a.readClientSecret()
	}
	return a
}

// readClientSecret : Read clienc secret
func (a *AuthContainer) readClientSecret() *AuthContainer {
	if csecret, err := ioutil.ReadFile(filepath.Join(a.InitVal.workdir, clientsecretFile)); err == nil {
		err := json.Unmarshal(csecret, &a.Cs)
		if err != nil || (len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) == 0) {
			fmt.Fprintf(os.Stderr, "Error: Please confirm '%s'. Error is %s.", clientsecretFile, err)
			os.Exit(1)
		}
		if len(a.Cs.Cid.ClientID) == 0 && len(a.Cs.Ciw.ClientID) > 0 {
			a.Cs.Cid = a.Cs.Ciw
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: No materials for retrieving accesstoken. Please download '%s'", clientsecretFile)
		os.Exit(1)
	}
	return a
}

// goauth : Main of auth process
func (a *AuthContainer) goauth() *AuthContainer {
	if len(a.gogauthCfg.Clientid) > 0 &&
		len(a.gogauthCfg.Clientsecret) > 0 &&
		len(a.gogauthCfg.Refreshtoken) > 0 {
		if (a.InitVal.pstart.Unix()-a.gogauthCfg.Expiresin) > 0 ||
			len(a.gogauthCfg.Accesstoken) == 0 {
			a.getAtoken().makecfgfile()
		} else {
			if a.InitVal.update {
				a.makecfgfile()
			}
		}
	} else {
		a.readClientSecret().getNewAccesstoken().makecfgfile()
	}
	return a
}

// reAuth : Reget refreshtoken
func (a *AuthContainer) reAuth() {
	a.readClientSecret().getNewAccesstoken().makecfgfile()
}

// makecfgfile : Write a cfg file
func (a *AuthContainer) makecfgfile() {
	btok, _ := json.MarshalIndent(a.gogauthCfg, "", "\t")
	ioutil.WriteFile(filepath.Join(a.InitVal.workdir, cfgFile), btok, 0777)
}

// getAtoken : Retrieves accesstoken from refreshtoken.
func (a *AuthContainer) getAtoken() *AuthContainer {
	values := url.Values{}
	values.Set("client_id", a.gogauthCfg.Clientid)
	values.Set("client_secret", a.gogauthCfg.Clientsecret)
	values.Set("refresh_token", a.gogauthCfg.Refreshtoken)
	values.Set("grant_type", "refresh_token")
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      oauthurl + "token",
		Data:        strings.NewReader(values.Encode()),
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: "",
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v. ", err)
		os.Exit(1)
	}
	json.Unmarshal(body, &a.Atoken)
	a.gogauthCfg.Accesstoken = a.Atoken.Accesstoken
	a.gogauthCfg.Expiresin = a.chkAtoken() - 360 // 6 minutes as adjustment time
	return a
}

// chkAtoken : For AuthContainer
func (a *AuthContainer) chkAtoken() int64 {
	r := &utl.RequestParams{
		Method:      "GET",
		APIURL:      chkatutl + "tokeninfo?access_token=" + a.gogauthCfg.Accesstoken,
		Data:        nil,
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: "",
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v%v.\n", string(body), err)
		os.Exit(1)
	}
	json.Unmarshal(body, &a.chkAt)
	if len(a.chkAt.Error) > 0 {
		a.getAtoken()
	}
	exp, _ := strconv.ParseInt(a.chkAt.Exp, 10, 64)
	return exp
}

// chkRedirectURI : Check redirect URI
func (a *AuthContainer) chkRedirectURI() bool {
	for _, e := range a.Cs.Cid.Redirecturis {
		if strings.Contains(e, "localhost") {
			return true
		}
	}
	return false
}

// getCode : Retrieve code on browser
func (a *AuthContainer) getCode() (string, error) {
	p := a.InitVal.Port
	if !a.chkRedirectURI() {
		return "", fmt.Errorf("Go manual mode.")
	}
	fmt.Printf("\n### This is a automatic input mode.\n### Please follow opened browser, login Google and click authentication.\n### Moves to a manual mode if you wait for 30 seconds under this situation.\n")
	a.Cs.Cid.Redirecturis = append(a.Cs.Cid.Redirecturis, "http://localhost:"+strconv.Itoa(p)+"/")
	codepara := url.Values{}
	codepara.Set("client_id", a.Cs.Cid.ClientID)
	codepara.Set("redirect_uri", a.Cs.Cid.Redirecturis[len(a.Cs.Cid.Redirecturis)-1])
	codepara.Set("scope", strings.Join(a.gogauthCfg.Scopes, " "))
	codepara.Set("response_type", "code")
	codepara.Set("approval_prompt", "force")
	codepara.Set("access_type", "offline")
	codeurl := oauthurl + "auth?" + codepara.Encode()
	s := &serverInfToGetCode{
		Response: make(chan authCode, 1),
		Start:    make(chan bool, 1),
		End:      make(chan bool, 1),
	}
	defer func() {
		s.End <- true
	}()
	go func(port int) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if len(code) == 0 {
				fmt.Fprintf(w, `<html><head><title>gogauth status</title></head><body><p>Erorr.</p></body></html>`)
				s.Response <- authCode{Err: fmt.Errorf("Not found code.")}
				return
			}
			fmt.Fprintf(w, `<html><head><title>gogauth status</title></head><body><p>The authentication was done. Please close this page.</p></body></html>`)
			s.Response <- authCode{Code: code}
		})
		var err error
		Listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			s.Response <- authCode{Err: err}
			return
		}
		server := http.Server{}
		server.Handler = mux
		go server.Serve(Listener)
		s.Start <- true
		<-s.End
		Listener.Close()
		s.Response <- authCode{Err: err}
		return
	}(p)
	<-s.Start
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", strings.Replace(codeurl, "&", `\&`, -1))
	case "linux":
		cmd = exec.Command("xdg-open", strings.Replace(codeurl, "&", `\&`, -1))
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", strings.Replace(codeurl, "&", `^&`, -1))
	default:
		return "", fmt.Errorf("Go manual mode.")
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("Go manual mode.")
	}
	var result authCode
	select {
	case result = <-s.Response:
	case <-time.After(time.Duration(30) * time.Second): // After 30 s, move to manual mode.
		return "", fmt.Errorf("Go manual mode.")
	}
	if result.Err != nil {
		return "", fmt.Errorf("Go manual mode.")
	}
	return result.Code, nil
}

// getNewAccesstoken : Retrieve accesstoken when there is no refreshtoken.
func (a *AuthContainer) getNewAccesstoken() *AuthContainer {
	var code string
	var err error
	code, err = a.getCode()
	if err != nil {
		codepara := url.Values{}
		codepara.Set("client_id", a.Cs.Cid.ClientID)
		codepara.Set("redirect_uri", a.Cs.Cid.Redirecturis[0])
		codepara.Set("scope", strings.Join(a.gogauthCfg.Scopes, " "))
		codepara.Set("response_type", "code")
		codepara.Set("approval_prompt", "force")
		codepara.Set("access_type", "offline")
		codeurl := oauthurl + "auth?" + codepara.Encode()
		fmt.Printf("\n### This is a manual input mode.\n### Please input code retrieved by importing following URL to your browser.\n\n"+
			"[URL]==> %v\n"+
			"[CODE]==>", codeurl)
		if _, err := fmt.Scan(&code); err != nil {
			log.Fatalf("Error: %v.\n", err)
		}
		a.Cs.Cid.Redirecturis = append(a.Cs.Cid.Redirecturis, a.Cs.Cid.Redirecturis[0])
	}
	tokenparams := url.Values{}
	tokenparams.Set("client_id", a.Cs.Cid.ClientID)
	tokenparams.Set("client_secret", a.Cs.Cid.Clientsecret)
	tokenparams.Set("redirect_uri", a.Cs.Cid.Redirecturis[len(a.Cs.Cid.Redirecturis)-1])
	tokenparams.Set("code", code)
	tokenparams.Set("grant_type", "authorization_code")
	r := &utl.RequestParams{
		Method:      "POST",
		APIURL:      oauthurl + "token",
		Data:        strings.NewReader(tokenparams.Encode()),
		Contenttype: "application/x-www-form-urlencoded",
		Accesstoken: "",
		Dtime:       10,
	}
	body, err := r.FetchAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: [ %v ] - Code is wrong. ", err)
		os.Exit(1)
	}
	json.Unmarshal(body, &a.Atoken)
	a.gogauthCfg.Clientid = a.Cs.Cid.ClientID
	a.gogauthCfg.Clientsecret = a.Cs.Cid.Clientsecret
	a.gogauthCfg.Refreshtoken = a.Atoken.Refreshtoken
	a.gogauthCfg.Accesstoken = a.Atoken.Accesstoken
	a.gogauthCfg.Expiresin = a.chkAtoken() - 360 // 6 minutes as adjustment time
	return a
}
