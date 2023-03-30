package auth

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type AuthType string

const (
	SSHAgentAuth AuthType = "ssh-agent"
	DefaultAuth  AuthType = "default"
)

type Auth struct {
	auth transport.AuthMethod
}

func NewAuth(authType AuthType) (*Auth, error) {
	var err error
	a := &Auth{}
	switch authType {
	case SSHAgentAuth:
		a.auth, err = ssh.NewSSHAgentAuth("git")
	case DefaultAuth:
		a.auth = nil
	default:
		err = fmt.Errorf("auth type %s not supported. valid values: %s, %s", authType, SSHAgentAuth, DefaultAuth)
	}
	return a, err
}

func (a Auth) Get() transport.AuthMethod {
	return a.auth
}
