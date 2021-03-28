package groomba

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestMoveBranchError(t *testing.T) {
	t.Run("MoveBranchError should return expected error output copy operation", func(t *testing.T) {
		a := assert.New(t)
		m := &MoveBranchError{branch: "errBranch1", operation: CopyBranch, err: fmt.Errorf("some error for errBranch1")}
		expectedErrMsg := "branch: errBranch1 failed on operation copy with error: some error for errBranch1"
		a.Equal(expectedErrMsg, m.Error())
	})
	t.Run("MoveBranchError should return expected error output delete operation", func(t *testing.T) {
		a := assert.New(t)
		m := &MoveBranchError{branch: "errBranch2", operation: DeleteBranch, err: fmt.Errorf("some other error for errBranch2")}
		expectedErrMsg := "branch: errBranch2 failed on operation delete with error: some other error for errBranch2"
		a.Equal(expectedErrMsg, m.Error())
	})
	t.Run("MoveBranchError should return expected error output and have copy operation as default", func(t *testing.T) {
		a := assert.New(t)
		m := &MoveBranchError{branch: "errBranch3", err: fmt.Errorf("some more errors for errBranch3")}
		expectedErrMsg := "branch: errBranch3 failed on operation copy with error: some more errors for errBranch3"
		a.Equal(expectedErrMsg, m.Error())
	})
}

func TestMoveStaleBranchesError(t *testing.T) {
	errList := []MoveBranchError{}
	errList = append(errList, MoveBranchError{branch: "errBranch1", operation: CopyBranch, err: fmt.Errorf("some error for errBranch1")})
	errList = append(errList, MoveBranchError{branch: "errBranch2", operation: DeleteBranch, err: fmt.Errorf("some other error for errBranch2")})
	errList = append(errList, MoveBranchError{branch: "errBranch3", err: fmt.Errorf("some more errors for errBranch3")})

	expectedErrMsg := "branch: errBranch1 failed on operation copy with error: some error for errBranch1\n" +
						"branch: errBranch2 failed on operation delete with error: some other error for errBranch2\n" +
						"branch: errBranch3 failed on operation copy with error: some more errors for errBranch3"

	m := MoveStaleBranchesError{errList: errList}
	t.Run("MoveStaleBranchesError should return the expected joined output", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(expectedErrMsg, m.Error())
	})
}

