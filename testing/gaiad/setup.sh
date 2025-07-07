#!/bin/bash

set -ex

# gaiad validator1
dir="${PWD}"

rm -rf "$dir/gaia1" || true
rm -rf "$dir/gaia2" || true
rm -rf "$dir/gaia3" || true

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


# collect gentxs
#

gaiad genesis add-genesis-account $(gaiad keys show validator3 -a --home $dir/gaia3 --keyring-backend test) 100000000000000000000stake --home $dir/gaia1
gaiad genesis add-genesis-account $(gaiad keys show validator2 -a --home $dir/gaia2 --keyring-backend test) 100000000000000000000stake --home $dir/gaia1



cp ${dir}/gaia3/config/gentx/* ${dir}/gaia1/config/gentx/
cp ${dir}/gaia2/config/gentx/* ${dir}/gaia1/config/gentx/
gaiad genesis collect-gentxs --home "$dir/gaia1"



cp "$dir/gaia3/config/priv_validator_key.json" "/home/shawn/github/cometkms/priv_validator_key.json"
