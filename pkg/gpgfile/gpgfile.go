package gpgfile

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/PuerkitoBio/purell"
	"github.com/pkg/errors"
)

const defaultFilePerms = 0600

// purellFlags are used to normalize VAULT_ADDR using the purell lib
const purellFlags = purell.FlagsSafe | purell.FlagsUsuallySafeGreedy | purell.FlagRemoveDuplicateSlashes

type tokenStore struct {
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

func New(file, gpgKey string) (tokenStore, error) {
	store := tokenStore{
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

func (s tokenStore) Get(vaultAddr string) string {
	vaultAddr = normalizeVaultAddr(vaultAddr)
	if token, exists := s.store[vaultAddr]; exists {
		return token
	}
	return ""
}

func (s tokenStore) Store(vaultAddr, token string) error {
	vaultAddr = normalizeVaultAddr(vaultAddr)
	s.store[vaultAddr] = token
	return s.encryptFile()
}

func (s tokenStore) Erase(vaultAddr string) error {
	vaultAddr = normalizeVaultAddr(vaultAddr)
	delete(s.store, vaultAddr)
	return s.encryptFile()
}

func (s tokenStore) decryptFile() ([]byte, error) {
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

func (s tokenStore) encryptFile() error {
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

func normalizeVaultAddr(addr string) string {
	normalized, err := purell.NormalizeURLString(addr, purellFlags)
	if err != nil {
		return addr
	}
	return normalized
}
