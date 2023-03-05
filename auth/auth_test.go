package auth

import (
	"fmt"
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
		a, err := NewAuth(SSHAgentAuth)
		assert.Nil(t, err)
		_, ok := a.auth.(*ssh.PublicKeysCallback)
		assert.True(t, ok, "expected returned auth type to be of ssh-agent type")

	})

	t.Run("default auth", func(t *testing.T) {
		u := AuthType("undefined-auth-type")
		a, err := NewAuth(u)
		assert.EqualError(t, err, fmt.Sprintf("auth type %s not supported. valid values: %s, %s", u, SSHAgentAuth, DefaultAuth))
		assert.Nil(t, a.auth)
	})

}
