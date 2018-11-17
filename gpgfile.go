package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

const defaultFilePerms = 0600

type gpgTokenStore struct {
	file   string
	gpgKey string
	store  map[string]string
}

func gpgBin() string {
	bin := "gpg"
	if envBin := os.Getenv("VAULT_GPG_BIN"); envBin != "" {
		bin = envBin
	}
	return bin
}

func newGPGTokenStore(file, gpgKey string) (gpgTokenStore, error) {
	store := gpgTokenStore{
		file:   file,
		gpgKey: gpgKey,
		store:  make(map[string]string),
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		// return a new empty store
		return store, nil
	}

	contents, err := store.decryptFile()
	if err != nil {
		return store, errors.Wrapf(err, "Failed to open token storage file (%s)", file)
	}
	if err := json.Unmarshal(contents, &store.store); err != nil {
		return store, errors.Wrapf(err, "Failed to parse token storage file (%s)", file)
	}

	return store, nil
}

func (s gpgTokenStore) Get(vaultAddr string) string {
	if token, exists := s.store[vaultAddr]; exists {
		return token
	}
	return ""
}

func (s gpgTokenStore) Store(vaultAddr, token string) error {
	s.store[vaultAddr] = token
	return s.encryptFile()
}

func (s gpgTokenStore) Erase(vaultAddr string) error {
	delete(s.store, vaultAddr)
	return s.encryptFile()
}

func (s gpgTokenStore) decryptFile() ([]byte, error) {
	encrypted, err := ioutil.ReadFile(s.file)
	if err != nil {
		return nil, err
	}

	args := []string{
		"--batch",
		"--use-agent",
		"-d",
	}
	cmd := exec.Command(gpgBin(), args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdin = bytes.NewReader(encrypted)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "GPG decryption failed: %s", stderr.String())
	}

	return stdout.Bytes(), nil
}

func (s gpgTokenStore) encryptFile() error {
	contents, err := json.Marshal(s.store)
	if err != nil {
		return err
	}

	args := []string{
		"--batch",
		"--no-default-recipient",
		"--yes",
		"--encrypt",
		"-a",
		"-r",
		s.gpgKey,
		// "--trusted-key",
		// s.gpgKey,
		"--no-encrypt-to",
	}
	cmd := exec.Command(gpgBin(), args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdin = bytes.NewReader(contents)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "GPG encryption failed: %s", stderr.String())
	}

	return ioutil.WriteFile(s.file, stdout.Bytes(), defaultFilePerms)
}
