package groomba

/*
   Copyright 2021 Amod Mulay

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"fmt"
	"strings"
)

// MoveBranchOperation defines the various operatons during MoveBranch
type MoveBranchOperation int

const (
	CopyBranch MoveBranchOperation = iota
	DeleteBranch
)

// MoveBranchError defines the errors during the MoveBranch step
type MoveBranchError struct {
	branch    string
	operation MoveBranchOperation
	err       error
}

// Error so MoveBranchError satisfies the error interface
func (e *MoveBranchError) Error() string {
	if e.operation == CopyBranch {
		return fmt.Sprintf("branch: %s failed on operation copy with error: %s", e.branch, e.err)
	}
	return fmt.Sprintf("branch: %s failed on operation delete with error: %s", e.branch, e.err)
}

// Unwrap for MoveBranchError
func (e *MoveBranchError) Unwrap() error {
	return e.err
}

// MoveStaleBranchesError stores all errors from MoveBranches
type MoveStaleBranchesError struct {
	errList []MoveBranchError
}

// Error so MoveStaleBranchesError satisfies the error interface
func (m *MoveStaleBranchesError) Error() string {
	msgList := []string{}
	for _, err := range m.errList {
		msgList = append(msgList, err.Error())
	}
	return strings.Join(msgList, "\n")
}
