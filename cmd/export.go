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
	"archive/tar"
	"errors"
	"fmt"
	"os"

	"github.com/wulonghui/ca-ctl/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/wulonghui/ca-ctl/depot"
	"github.com/wulonghui/ca-ctl/pkix"
)

const (
	crtSuffix      = ".crt"
	keySuffix      = ".key"
	insecureSuffix = ".insecure"
)

func NewExportCommand() cli.Command {
	return cli.Command{
		Name:        "export",
		Usage:       "Export host certificate and key",
		Description: "Package up a certificate and key for export to a server. Without args, it exports CA certificate and key.",
		Flags: []cli.Flag{
			cli.BoolFlag{"insecure", "Export private key without encryption", ""},
			cli.StringFlag{"passphrase", "", "Passphrase to decrypt private-key PEM block", ""},
		},
		Action: newExportAction,
	}
}

type TarFile struct {
	Header *tar.Header
	Data   []byte
}

func newExportAction(c *cli.Context) {
	if len(c.Args()) > 1 {
		fmt.Fprintln(os.Stderr, "At most one host name could be provided.")
		os.Exit(1)
	}

	if _, err := depot.GetCertificateAuthority(d); isFileNotExist(err) {
		fmt.Fprintln(os.Stderr, "Please run 'ca-ctl init' to initial the depot.")
		os.Exit(1)
	}

	var files []*TarFile
	var err error
	if len(c.Args()) == 0 {
		files, err = getAuthFiles(c)
	} else {
		files, err = getHostFiles(c, c.Args()[0])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w := tar.NewWriter(os.Stdout)
	defer w.Close()
	if err = outputTarFiles(w, files); err != nil {
		fmt.Fprintln(os.Stderr, "Save tar error:", err)
		os.Exit(1)
	}
}

func getAuthFiles(c *cli.Context) ([]*TarFile, error) {
	name := "ca"
	tarFiles := make([]*TarFile, 0)

	crtFile, err := d.GetFile(depot.AuthCrtTag())
	if err != nil {
		return nil, errors.New("Get CA certificate error: " + err.Error())
	}
	crtTarFile, err := generateTarFile(crtFile, name+crtSuffix)
	if err != nil {
		return nil, errors.New("Generate certificate tar file error: " + err.Error())
	}
	tarFiles = append(tarFiles, crtTarFile)

	keyFile, err := d.GetFile(depot.AuthPrivKeyTag())
	if err != nil {
		return nil, errors.New("Get CA key error: " + err.Error())
	}
	keyTarFile, err := generateTarFile(keyFile, name+keySuffix)
	if err != nil {
		return nil, errors.New("Generate key tar file error: " + err.Error())
	}
	if c.Bool("insecure") {
		if keyTarFile, err = decryptEncryptedKeyTarFile(keyTarFile, getPassPhrase(c, name+" key")); err != nil {
			return nil, errors.New("Get decrypted CA key error: " + err.Error())
		}
	}
	tarFiles = append(tarFiles, keyTarFile)

	return tarFiles, nil
}

func getHostFiles(c *cli.Context, name string) ([]*TarFile, error) {
	tarFiles := make([]*TarFile, 0)

	crtFile, err := d.GetFile(depot.HostCrtTag(name))
	if err != nil {
		return nil, errors.New("Get host certificate error: " + err.Error())
	}
	crtTarFile, err := generateTarFile(crtFile, name+crtSuffix)
	if err != nil {
		return nil, errors.New("Generate certificate tar file error: " + err.Error())
	}
	tarFiles = append(tarFiles, crtTarFile)

	keyFile, err := d.GetFile(depot.HostPrivKeyTag(name))
	if err != nil {
		return nil, errors.New("Get host key error: " + err.Error())
	}
	keyTarFile, err := generateTarFile(keyFile, name+keySuffix)
	if err != nil {
		return nil, errors.New("Generate key tar file error: " + err.Error())
	}
	if c.Bool("insecure") {
		if keyTarFile, err = decryptEncryptedKeyTarFile(keyTarFile, getPassPhrase(c, name+" key")); err != nil {
			return nil, errors.New("Get decrypted host key error: " + err.Error())
		}
	}
	tarFiles = append(tarFiles, keyTarFile)

	return tarFiles, nil
}

func decryptEncryptedKeyTarFile(file *TarFile, passphrase []byte) (*TarFile, error) {
	key, err := pkix.NewKeyFromEncryptedPrivateKeyPEM(file.Data, passphrase)
	if err != nil {
		return nil, err
	}
	file.Data, err = key.ExportPrivate()
	if err != nil {
		return nil, err
	}
	file.Header.Name += insecureSuffix
	file.Header.Size = int64(len(file.Data))
	return file, nil
}

func generateTarFile(file *depot.File, newName string) (*TarFile, error) {
	header, err := tar.FileInfoHeader(file.Info, "")
	if err != nil {
		return nil, err
	}
	header.Name = newName
	header.Mode |= 0644
	return &TarFile{header, file.Data}, nil
}

func outputTarFiles(out *tar.Writer, files []*TarFile) error {
	for _, file := range files {
		if err := out.WriteHeader(file.Header); err != nil {
			return err
		}
		if _, err := out.Write(file.Data); err != nil {
			return err
		}
	}

	return nil
}
