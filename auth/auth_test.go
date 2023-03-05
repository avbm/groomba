package auth

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/stretchr/testify/assert"
)

func TestNewAuth(t *testing.T) {

	t.Run("default auth", func(t *testing.T) {
		a, err := NewAuth(DefaultAuth)
		assert.Nil(t, err)
		assert.Nil(t, a.auth)
	})

	t.Run("ssh-agent auth", func(t *testing.T) {
		// First test that we return error when trying to use ssh-agent auth with no agent running
		sockVar := "SSH_AUTH_SOCK"
		pidVar := "SSH_AGENT_PID"
		if agentSock := os.Getenv(sockVar); agentSock != "" {
			defer os.Setenv(sockVar, agentSock)
		}
		if agentPID := os.Getenv(pidVar); agentPID != "" {
			defer os.Setenv(pidVar, agentPID)
		}
		os.Unsetenv(sockVar)
		os.Unsetenv(pidVar)
		a, err := NewAuth(SSHAgentAuth)
		assert.EqualError(t, err, `error creating SSH agent: "SSH agent requested but SSH_AUTH_SOCK not-specified"`)
		_, ok := a.auth.(*ssh.PublicKeysCallback)
		assert.True(t, ok, "expected returned auth type to be of ssh-agent type")

		// Setup the ssh-agent
		stdout := &bytes.Buffer{}
		cmd := exec.Command("ssh-agent", "-s")
		cmd.Stdout = stdout
		err = cmd.Run()
		assert.Nil(t, err, "failed to setup ssh-agent for test")
		for _, s := range strings.Split(stdout.String(), ";") {
			if strings.Contains(s, "=") {
				envVar := strings.Split(s, "=")
				os.Setenv(envVar[0], envVar[1])
			}
		}

		// Ensure ssh-agent authenticator initialization works with ssh-agent running
		a, err = NewAuth(SSHAgentAuth)
		assert.Nil(t, err)
		_, ok = a.auth.(*ssh.PublicKeysCallback)
		assert.True(t, ok, "expected returned auth type to be of ssh-agent type")

	})

	t.Run("default auth", func(t *testing.T) {
		u := AuthType("undefined-auth-type")
		a, err := NewAuth(u)
		assert.EqualError(t, err, fmt.Sprintf("auth type %s not supported. valid values: %s, %s", u, SSHAgentAuth, DefaultAuth))
		assert.Nil(t, a.auth)
	})

}
