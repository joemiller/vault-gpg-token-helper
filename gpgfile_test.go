package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

const (
	testStore  = "./test-store.gpg"
	testKeyID  = "test-key" // this must match the "Name-Real:" in the gpgConfig
	gpgTestDir = "./gpghometest"

	gpgConfig = `
%echo Generating a default key
Key-Type: default
Subkey-Type: default
Name-Real: test-key
Name-Comment: testing
Name-Email: joe@foo.bar
Expire-Date: 0
%no-protection
# %pubring test.pub
# %secring test.sec
# Do a commit here, so that we can later print "done" :-)
%commit
%echo done
`
)

// setupGpg creates a temporary GNUPGHOME directory and initializes it with a key with I 'test-key'
func setupGpg() error {
	gpgHome := "./gpghometest"
	if err := os.Mkdir(gpgHome, 0700); err != nil {
		return err
	}

	if err := os.Setenv("GNUPGHOME", "./gpghometest"); err != nil {
		return err
	}
	args := []string{
		"--batch",
		"--quiet",
		"--gen-key",
		"--debug-quick-random",
	}
	cmd := exec.Command(gpgBin(), args...)
	cmd.Stdin = strings.NewReader(gpgConfig)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, string(out))
	}
	return nil
}

func cleanup(t *testing.T) {
	if os.Getenv("SKIP_TEST_CLEANUP") != "" {
		return
	}
	if err := os.RemoveAll(gpgTestDir); err != nil {
		t.Log(err)
	}
	if err := os.Remove(testStore); err != nil {
		t.Log(err)
	}
}

func TestGPGTokenStore(t *testing.T) {
	if err := setupGpg(); err != nil {
		t.Fatal(err)
	}
	defer cleanup(t)

	store, err := newGPGTokenStore(testStore, testKeyID)
	if err != nil {
		t.Fatal("failed to initialize new GPG token store: ", err)
	}

	// new store should be empty
	if len(store.store) != 0 {
		t.Error("expected store to be empty")
	}

	// add an entry
	if err := store.Store("https://vault1:8200", "token-foo"); err != nil {
		t.Error("unexpected error storing token: ", err)
	}

	// read back the entry
	token := store.Get("httpS://VaULT1:8200//")
	if token != "token-foo" {
		t.Errorf("expected 'token-foo' got '%s'", token)
	}

	// erase the entry
	if err := store.Erase("HTTPS://vault1:8200"); err != nil {
		t.Error("unexpeted error erasing token from the store: ", err)
	}

	// read back the entry, it should be gone
	token = store.Get("https://vault1:8200/")
	if token != "" {
		t.Errorf("expected '' got '%s'", token)
	}
	//spew.Dump(store)
}
