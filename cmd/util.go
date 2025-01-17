// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/wulonghui/ca-ctl/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/wulonghui/ca-ctl/Godeps/_workspace/src/golang.org/x/crypto/ssh/terminal"
	"github.com/wulonghui/ca-ctl/depot"
)

var (
	d *depot.FileDepot
)

func InitDepot(path string) error {
	if d == nil {
		var err error
		if d, err = depot.NewFileDepot(path); err != nil {
			return err
		}
	}
	return nil
}

func createPassPhrase() ([]byte, error) {
	fmt.Fprint(os.Stderr, "Enter passphrase (empty for no passphrase): ")
	pass1, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return nil, err
	}
	fmt.Fprint(os.Stderr, "\nEnter same passphrase again: ")
	pass2, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stderr)

	if bytes.Compare(pass1, pass2) != 0 {
		return nil, errors.New("Passphrases do not match.")
	}
	return pass1, nil
}

func askPassPhrase(name string) []byte {
	fmt.Fprintf(os.Stderr, "Enter passphrase for %v (empty for no passphrase): ", name)
	pass, _ := terminal.ReadPassword(syscall.Stdin)
	fmt.Fprintln(os.Stderr)
	return pass
}

func getPassPhrase(c *cli.Context, name string) []byte {
	if c.IsSet("passphrase") {
		return []byte(c.String("passphrase"))
	} else {
		return askPassPhrase(name)
	}
}

func isFileNotExist(err error) bool {
	perr, ok := err.(*os.PathError)
	return ok && perr.Err.Error() == "no such file or directory"
}
