#!/bin/bash

set -ex

# gaiad validator1
dir="${PWD}"

rm -rf "$dir/gaia1" || true
rm -rf "$dir/gaia2" || true
rm -rf "$dir/gaia3" || true
rm -rf "$dir/gaia4" || true

if [ -d "$dir/gaia1" ]; then
  rm -r "$dir/gaia1"
else
  echo "Directory $dir/gaia1 does not exist."
fi

gaiad init localnet --chain-id=localnet --home "$dir/gaia1"
gaiad config keyring-backend test --home "$dir/gaia1"
gaiad keys add validator1 --home $dir/gaia1 --keyring-backend test
gaiad genesis add-genesis-account $(gaiad keys show validator1 -a --home $dir/gaia1 --keyring-backend test) 1000000000000000000000stake --home $dir/gaia1
gaiad genesis gentx validator1 100000000000000000stake --chain-id localnet --home $dir/gaia1 --keyring-backend test
sed -i "s|^minimum-gas-prices = \"\"|minimum-gas-prices = \"0stake\"|" $dir/gaia1/config/app.toml

sed -i 's/"signed_blocks_window": *"100"/"signed_blocks_window": "10000000000000"/' $dir/gaia1/config/genesis.json

# gaiad validator2
if [ -d "$dir/gaia2" ]; then
  rm -r "$dir/gaia2"
else
  echo "Directory $dir/gaia2 does not exist."
fi

gaiad init localnet --chain-id=localnet --home "$dir/gaia2"
gaiad config keyring-backend test --home "$dir/gaia2"
gaiad keys add validator2 --home $dir/gaia2 --keyring-backend test
gaiad genesis add-genesis-account $(gaiad keys show validator2 -a --home $dir/gaia2 --keyring-backend test) 1000000000000000000stake --home $dir/gaia2
gaiad genesis gentx validator2 100000000000000000stake --chain-id localnet --home $dir/gaia2 --keyring-backend test
sed -i "s|^minimum-gas-prices = \"\"|minimum-gas-prices = \"0stake\"|" $dir/gaia2/config/app.toml

# gaiad validator3

if [ -d "$dir/gaia3" ]; then
  rm -r "$dir/gaia3"
else
  echo "Directory $dir/gaia3 does not exist."
fi

gaiad init localnet --chain-id=localnet --home "$dir/gaia3"
gaiad config keyring-backend test --home "$dir/gaia3"
gaiad keys add validator3 --home $dir/gaia3 --keyring-backend test
gaiad genesis add-genesis-account $(gaiad keys show validator3 -a --home $dir/gaia3 --keyring-backend test) 10000000stake --home $dir/gaia3
gaiad genesis gentx validator3 1000000stake --chain-id localnet --home $dir/gaia3 --keyring-backend test
sed -i "s|^minimum-gas-prices = \"\"|minimum-gas-prices = \"0stake\"|" $dir/gaia3/config/app.toml
sed -i 's|^priv_validator_laddr = ""|priv_validator_laddr = "tcp://0.0.0.0:12345"|' $dir/gaia3/config/config.toml

# gaiad validator4
if [ -d "$dir/gaia4" ]; then
  rm -r "$dir/gaia4"
else
  echo "Directory $dir/gaia4 does not exist."
fi
gaiad init localnet --chain-id=localnet --home "$dir/gaia4"
gaiad config keyring-backend test --home "$dir/gaia4"
gaiad keys add validator4 --home $dir/gaia4 --keyring-backend test
gaiad genesis add-genesis-account $(gaiad keys show validator4 -a --home $dir/gaia4 --keyring-backend test) 10000000stake --home $dir/gaia4
gaiad genesis gentx validator4 1000000stake --chain-id localnet --home $dir/gaia4 --keyring-backend test
sed -i "s|^minimum-gas-prices = \"\"|minimum-gas-prices = \"0stake\"|" $dir/gaia4/config/app.toml
sed -i 's|^priv_validator_laddr = ""|priv_validator_laddr = "tcp://0.0.0.0:54321"|' $dir/gaia4/config/config.toml

# collect gentxs

gaiad genesis add-genesis-account $(gaiad keys show validator3 -a --home $dir/gaia3 --keyring-backend test) 100000000000000000000stake --home $dir/gaia1
gaiad genesis add-genesis-account $(gaiad keys show validator2 -a --home $dir/gaia2 --keyring-backend test) 100000000000000000000stake --home $dir/gaia1


cp ${dir}/gaia3/config/gentx/* ${dir}/gaia1/config/gentx/
cp ${dir}/gaia2/config/gentx/* ${dir}/gaia1/config/gentx/
gaiad genesis collect-gentxs --home "$dir/gaia1"

# copy genesis.json to other nodes
cp "$dir/gaia1/config/genesis.json" "$dir/gaia2/config/genesis.json"
cp "$dir/gaia1/config/genesis.json" "$dir/gaia3/config/genesis.json"
cp "$dir/gaia1/config/genesis.json" "$dir/gaia4/config/genesis.json"

# setup persistent_peers
peer1="$(gaiad tendermint show-node-id --home $dir/gaia1)"
peer2="$(gaiad tendermint show-node-id --home $dir/gaia2)"
peer3="$(gaiad tendermint show-node-id --home $dir/gaia3)"
peer4="$(gaiad tendermint show-node-id --home $dir/gaia4)"

port2=36656
port3=46656
port4=56656


# change port 9090 for each node
sed -i 's/9090/9091/' $dir/gaia2/config/app.toml
sed -i 's/9090/9092/' $dir/gaia3/config/app.toml
sed -i 's/9090/9093/' $dir/gaia4/config/app.toml

#change port 26657 for each node
sed -i 's/26657/36657/' $dir/gaia2/config/config.toml
sed -i 's/26657/46657/' $dir/gaia3/config/config.toml
sed -i 's/26657/56657/' $dir/gaia4/config/config.toml

#change port 6060
sed -i 's/6060/6061/' $dir/gaia2/config/config.toml
sed -i 's/6060/6062/' $dir/gaia3/config/config.toml
sed -i 's/6060/6063/' $dir/gaia4/config/config.toml

# change port 26656
sed -i 's/26656/36656/' $dir/gaia2/config/config.toml
sed -i 's/26656/46656/' $dir/gaia3/config/config.toml
sed -i 's/26656/56656/' $dir/gaia4/config/config.toml

sed -i "s|^persistent_peers = \"\"|persistent_peers = \"$peer1@127.0.0.1:26656,$peer2@127.0.0.1:$port2,$peer3@127.0.0.1:$port3,$peer4@127.0.0.1:$port4\"|" $dir/gaia1/config/config.toml
sed -i "s|^persistent_peers = \"\"|persistent_peers = \"$peer1@127.0.0.1:26656,$peer2@127.0.0.1:$port2,$peer3@127.0.0.1:$port3,$peer4@127.0.0.1:$port4\"|" $dir/gaia2/config/config.toml
sed -i "s|^persistent_peers = \"\"|persistent_peers = \"$peer1@127.0.0.1:26656,$peer2@127.0.0.1:$port2,$peer3@127.0.0.1:$port3,$peer4@127.0.0.1:$port4\"|" $dir/gaia3/config/config.toml
sed -i "s|^persistent_peers = \"\"|persistent_peers = \"$peer1@127.0.0.1:26656,$peer2@127.0.0.1:$port2,$peer3@127.0.0.1:$port3,$peer4@127.0.0.1:$port4\"|" $dir/gaia4/config/config.toml

# set pex to false
sed -i 's/pex = true/pex = false/' $dir/gaia1/config/config.toml
sed -i 's/pex = true/pex = false/' $dir/gaia2/config/config.toml
sed -i 's/pex = true/pex = false/' $dir/gaia3/config/config.toml
sed -i 's/pex = true/pex = false/' $dir/gaia4/config/config.toml

# allow duplicate ip
sed -i 's/allow_duplicate_ip = false/allow_duplicate_ip = true/' $dir/gaia1/config/config.toml
sed -i 's/allow_duplicate_ip = false/allow_duplicate_ip = true/' $dir/gaia2/config/config.toml
sed -i 's/allow_duplicate_ip = false/allow_duplicate_ip = true/' $dir/gaia3/config/config.toml
sed -i 's/allow_duplicate_ip = false/allow_duplicate_ip = true/' $dir/gaia4/config/config.toml

cp "$dir/gaia3/config/priv_validator_key.json" "/home/shawn/github/cometkms/priv_validator_key.json"
