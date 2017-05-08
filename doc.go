/*
Package main (doc.go) :
This is a CLI tool to retrieve easily access token for using APIs on Google.

# Features of this CLI tool is as follows.

1. Retrieves easily access token for using APIs on Google.

2. Uses expiration time of access token.

3. Confirms condition of access token. For the access token, you can see expiration time, owner, scopes and client id.

---------------------------------------------------------------

# Usage
Help

$ gogauth --help

Retrieving access token

$ gogauth g

After 1st run of `$ gogauth g`, you can see a configuration file `gogauth.cfg` on the current working directory. The access token is retrieved with `https://www.googleapis.com/auth/drive.readonly` as an initial scope. When you want to change the scope, please modify the `gogauth.cfg`. Of course, you can write several scopes. After modified it, please execute below.

$ gogauth g -r

By this, new access token and refresh token for the modified scopes are retrieved, and updated `gogauth.cfg`. From this version, the code for retrieving refresh token is retrieved by web server on gogauth. The port for the server is 8080 as a default port number. If you want to change the port, please run below.

$ gogauth g -r -p [ port number ]

Also you can see the condition of access token using this tool.

$ gogauth c -a [ access token ]

You can see release page https://github.com/tanaikech/gogauth/releases

---------------------------------------------------------------
*/
package main
