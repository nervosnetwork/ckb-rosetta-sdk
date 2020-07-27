ckb rosetta sdk
===============

## Background

### [Rosetta](https://www.rosetta-api.org/)

> Coinbase Standard for Blockchain Interaction

- [Specifications](https://github.com/coinbase/rosetta-specifications)
- [API](https://github.com/coinbase/rosetta-specifications/blob/master/api.json)
- [Github](https://github.com/coinbase/rosetta-sdk-go)

## Build

```
git checkout https://github.com/nervosnetwork/ckb-rosetta-sdk.git
cd ckb-rosetta-sdk/server
go mod download
go build .
```

## Run

1. Run ckb-rich-node

    This step can refer to ckb-rich-node [README.md](https://github.com/ququzone/ckb-rich-node)
    
2. Run ckb-rosetta server

    ```
    cd ckb-rosetta-sdk/server
    nohup ./server &
    ```
