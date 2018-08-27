package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

var (
	defaultTokenPath  = "~/.vault_tokens.gpg"
	defaultConfigFile = "~/.vault-gpg-token-helper.toml"
)

type configuration struct {
	VaultAddr      string `hcl:"default_vault_addr"`
	VaultGPGKey    string `hcl:"gpg_key_id"`
	TokenStorePath string `hcl:"token_db_file"`
}

type tokenStorer interface {
	Get(vaultAddr string) string
	Store(vaultAddr, token string) error
	Erase(vaultAddr string)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: missing 'command'")
		exe, _ := os.Executable()
		fmt.Println("This app is not intended to be run by directly")
		fmt.Println("Place the following in your '$HOME/.vault' file and run 'vault':")
		fmt.Printf("token_helper = \"%s\"\n", exe)
		os.Exit(1)
	}
	cmd := os.Args[1]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(100)
	}

	fullTokenPath, err := homedir.Expand(cfg.TokenStorePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(200)
	}

	store, err := newGPGTokenStore(fullTokenPath, cfg.VaultGPGKey)
	if err != nil {
		fmt.Println(err)
		os.Exit(200)
	}

	switch cmd {
	case "get":
		if token := store.Get(cfg.VaultAddr); token != "" {
			fmt.Print(token)
		}
	case "store":
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("ERROR: Failed to read token from STDIN: ", err)
			os.Exit(1)
		}
		token := string(stdin)
		token = strings.TrimSuffix(token, "\n")
		if err := store.Store(cfg.VaultAddr, token); err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	case "erase":
		if err := store.Erase(cfg.VaultAddr); err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
	default:
		fmt.Printf("ERROR: Unknown command '%s'", cmd)
		os.Exit(4)
	}
}

// loadConfig returns a configuration struct with settings layered in order of preference:
// - defaults
// - config file (default: ~/.vault-gpg-token.toml)
// - environment vars (highest priority)
func loadConfig() (configuration, error) {
	// default config
	cfg := configuration{
		TokenStorePath: defaultTokenPath,
	}

	// attempt to load settings from config file if available
	cFile := defaultConfigFile
	if envFile := os.Getenv("VAULT_GPG_CONFIG"); envFile != "" {
		cFile = envFile
	}
	configFilePath, err := homedir.Expand(cFile)
	if err != nil {
		return cfg, err
	}
	if _, err := os.Stat(configFilePath); err == nil {
		contents, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return cfg, err
		}
		if err := hcl.Decode(&cfg, string(contents)); err != nil {
			return cfg, err
		}
	}

	// finally layer in environment vars, which have higher preference over defaults and config file
	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		cfg.VaultAddr = addr
	}

	if key := os.Getenv("VAULT_GPG_KEY_ID"); key != "" {
		cfg.VaultGPGKey = key
	}

	if path := os.Getenv("VAULT_GPG_TOKEN_STORE"); path != "" {
		cfg.TokenStorePath = path
	}

	// validation
	if cfg.VaultAddr == "" {
		return cfg, errors.New("no VAULT_ADDR environment variable set")
	}
	if cfg.VaultGPGKey == "" {
		return cfg, errors.New("no VAULT_GPG_KEY_ID environment variable set, and no 'gpg_key_id' in config file")
	}

	return cfg, nil
}
