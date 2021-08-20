/*
customized struct and related methods
*/
package internal

import "time"

const (
	LoginWeb    = "./HTTP/index.html"
	ProfileWeb  = "./HTTP/profile.html"
	HTTPPort    = "localhost:5500"
	TCPPort     = ":12345"
	ConnTimeOut = 5 * time.Second
	MaxConn     = 1000
	MinConn     = 100
	Prefix      = 12345
)

// Status: Login & Profile
type Login struct {
	Username string
	Password string
}

type UserInfo struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Profile  string `json:"profile"`
}

// Login Staus: Success & Failure
type LoginSuccess struct {
	Success  bool
	Username string
	Nickname string
	Profile  string
}

type LoginFailure struct {
	Success bool
	Reason  string
}

// Json
type JsonType struct {
	Type string `json:"type"`
}

type JsonLogin struct {
	JsonType
	Username string `json:"username"`
	Password string `json:"password"`
}

type JsonLoginRst struct {
	JsonType
	Result   bool     `json:"result"`
	Reason   string   `json:"reason"`
	UserInfo UserInfo `json:"userinfo"`
}

type JsonUpdate struct {
	JsonType
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Profile  string `json:"profile"`
}

type JsonUpdateRst struct {
	JsonType
	Result bool   `json:"result"`
	Reason string `json:"reason"`
}
