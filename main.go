package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

// var (
// 	version = "dev"
// 	commit  = "none"
// 	date    = "unknown"
// )

var (
	tokenPath = "~/.vault_tokens.gpg"
	gpgKeyID  = ""
	vaultAddr = ""
)

type tokenStorer interface {
	Get(vaultAddr string) string
	Store(vaultAddr, token string) error
	Erase(vaultAddr string)
}

func main() {
	if err := loadConfig(); err != nil {
		fmt.Println(err)
		os.Exit(100)
	}

	fullTokenPath, err := homedir.Expand(tokenPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(200)
	}
	store, err := newGPGTokenStore(fullTokenPath, gpgKeyID)
	if err != nil {
		fmt.Println(err)
		os.Exit(200)
	}

	// missing command, just exit
	if len(os.Args) < 2 {
		os.Exit(0)
	}

	cmd := os.Args[1]
	switch cmd {
	case "get":
		if token := store.Get(vaultAddr); token != "" {
			fmt.Println(token)
		}
	case "store":
		token, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("ERROR: Failed to read token from STDIN")
			os.Exit(1)
		}
		token = strings.TrimSuffix(token, "\n")
		if err := store.Store(vaultAddr, token); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	case "erase":
		if err := store.Erase(vaultAddr); err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
	default:
		fmt.Printf("ERROR: Unknown command '%s'", cmd)
		os.Exit(4)
	}
}

// TODO: support a config file, even just for vault_gpg_key for now
// TODO: support alternative tokenpath ?
func loadConfig() error {
	vaultAddr = os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		return errors.New("no VAULT_ADDR environment variable set")
	}

	gpgKeyID = os.Getenv("VAULT_GPG_KEY")
	if gpgKeyID == "" {
		return errors.New("no VAULT_GPG_KEY environment variable set")
	}

	return nil
}
