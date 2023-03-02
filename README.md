# RIGO Blockchain

## Prerequisite

### golang
Install golang v1.19+

### protoc (protobuf compiler)
```bash
brew install protobuf

...

protoc --version
libprotoc 3.21.12
```

### protoc-gen-go
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go

...

protoc-gen-go --version
protoc-gen-go v1.28.1
```

### mkdocs
*To update and deploy documents....*

## Sources

```bash
git clone https://github.com/rigochain/rigo-go.git
```

## Build

```bash
make
```

## Run RIGO
```bash
build/darwin/rigo init --chain_id local_test_net --priv_validator_secret 1234

...

build/darwin/rigo start
```

* RIGO's root directory is `$HOME/.rigo`.
* RIGO's config file is `$HOME/.rigo/config/config.toml`.
* Genesis file is `$HOME/.rigo/config/gebesus.json`.
* Validator private key file is  `$HOME/.rigo/config/priv_validator_key.json`.
* Initial wallet files are in `$HOME/.rigo/walkeys/`.
* To show private key, run `build/darwin/rigo show-wallet-key {Wallet Key Files}`.  
  (e.g. run `build/darwin/rigo show-wallet-key ~/.rigo/config/priv_validator_key.json`)