vault-gpg-token-helper
======================

[![CircleCI](https://circleci.com/gh/joemiller/vault-gpg-token-helper.svg?style=svg)](https://circleci.com/gh/joemiller/vault-gpg-token-helper)

A @hashicorp Vault token helper for storing tokens in a GPG encrypted file. Support
for GPG with YubiKey.

Requirements
============

* `vault`
* `gpg` (Tested with 2.2.x, likely compatible with 1.x and 2.1)

A `gpg` binary should be in your `$PATH`. An explicit path can be set with the
`VAULT_GPG_BIN` environment variable.

This program uses the gpg binary instead of Go's opengpg library to make it possible
to utilize GPG keys stored on a hardware device such as a YubiKey.

Install
=======

* Binary releases are [available](https://github.com/joemiller/vault-gpg-token-helper/releases) for many platforms.
* Homebrew (macOS): `brew install joemiller/taps/vault-gpg-token-helper`

After installation:

Create a `~/.vault` file with contents:

```toml
token_helper = "/path/to/vault-gpg-token-helper"
```

For homebrew installations you can create this file by running:

```shell
echo "token_helper = \"$(brew --prefix joemiller/taps/vault-gpg-token-helper)/bin/vault-gpg-token-helper\"" > ~/.vault
```

Configuration
=============

The default config file is `~/.vault-gpg-token-helper.toml`. This can be changed with the
`VAULT_GPG_CONFIG` environment variable.

At minimum a `gpg_key_id` must be set in the config file. Alternatively it can be
specified by the `VAULT_GPG_KEY_ID` environment variable.

Example:

```toml
gpg_key_id = "first last (yubikey) <firstlast@dom.tld>"
```

> Run `gpg --list-keys` for a list of keys.

## Token Storage

Tokens are stored encrypted in `~/.vault_tokens.gpg` by default. This can be
changed by:

* Setting the `token_db_file` configuration file option
* Setting the `VAULT_GPG_TOKEN_STORE` environment variable

Environment variables take precedence over configuration file settings.

Usage
=====

The `VAULT_ADDR` environment variable must be set. The storer uses this variable
as an index for storing and retrieving tokens. This allows for easy switching
between multiple Vault targets.

Example, adding a token to the store:

```shell
export VAULT_ADDR="https://vault-a:8200"
vault login
```

> Vault 0.10.2+ supports a `-no-print` flag to store the token without printing to stdout

Support
=======

Please open a GitHub issue.

TODO
====

- [ ] refactor to use the opengpg go lib. Ideally this lib would support
      yubikey, alternatively we could code it to use native code for software
      keys and fallback to shelling out to gpg for stubbed hardware keys. mozilla/sops
      project does this.
