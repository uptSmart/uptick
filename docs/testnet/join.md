<!--
order: 1
-->

# Joining a Testnet

This document outlines the steps to join an existing testnet {synopsis}

## Pick a Testnet

You specify the network you want to join by setting the **genesis file** and **seeds**. If you need more information about past networks, check our [testnets repo](https://github.com/UptickNetwork/uptick-testnet).

| Network Chain ID | Description                       | Site                                                                     | Version                                               |
|------------------|-----------------------------------|--------------------------------------------------------------------------|-------------------------------------------------------|
| `uptick_7777-1`   | Uptick Testnet | [uptick_7777-1 testnet](https://github.com/UptickNetwork/uptick-testnet/tree/main/uptick_7777-1) | [`v0.1.x`](https://github.com/UptickNetwork/uptick/releases) |

## Install `uptickd`

Follow the [installation](./../quickstart/installation) document to install the {{ $themeConfig.project.name }} binary `{{ $themeConfig.project.binary }}`.

:::warning
Make sure you have the right version of `{{ $themeConfig.project.binary }}` installed.
:::

### Save Chain ID

We recommend saving the mainnet `chain-id` into your `{{ $themeConfig.project.binary }}`'s `client.toml`. This will make it so you do not have to manually pass in the `chain-id` flag for every CLI command.

::: tip
See the Official [Chain IDs](./../basics/chain_id.md#official-chain-ids) for reference.
:::

```bash
uptickd config chain-id uptick_7777-1
```

## Initialize Node

We need to initialize the node to create all the necessary validator and node configuration files:

```bash
uptickd init <your_custom_moniker> --chain-id uptick_7777-1
```

::: danger
Monikers can contain only ASCII characters. Using Unicode characters will render your node unreachable.
:::

By default, the `init` command creates your `~/.uptickd` (i.e `$HOME`) directory with subfolders `config/` and `data/`.
In the `config` directory, the most important files for configuration are `app.toml` and `config.toml`.

## Genesis & Seeds

### Copy the Genesis File

Check the `genesis.json` file from the [`testnets`](https://github.com/UptickNetwork/uptick-testnet) repository and copy it over to the `config` directory: `~/.uptickd/config/genesis.json`. This is a genesis file with the chain-id and genesis accounts balances.

```bash
curl https://raw.githubusercontent.com/UptickNetwork/uptick-testnet/main/uptick_7777-1/genesis.json > ~/.uptickd/config/genesis.json
```

Then verify the correctness of the genesis configuration file:

```bash
uptickd validate-genesis
```

### Add Seed Nodes

Your node needs to know how to find [peers](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#peers). You'll need to add healthy [seed nodes](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#seed) to `$HOME/.uptickd/config/config.toml`. The [`testnets`](https://github.com/UptickNetwork/uptick-testnet) repo contains links to some seed nodes.

Edit the file located in `~/.uptickd/config/config.toml` and the `seeds` to the following:

```toml
#######################################################
###           P2P Configuration Options             ###
#######################################################
[p2p]

# ...

# Comma separated list of seed nodes to connect to
seeds = "<node-id>@<ip>:<p2p port>"
```

You can use the following code to get seeds from the repo and add it to your config:

```bash
SEEDS=`curl -sL https://raw.githubusercontent.com/UptickNetwork/uptick-testnet/main/uptick_7777-1/seeds.txt | awk '{print $1}' | paste -s -d, -`
sed -i.bak -e "s/^seeds =.*/seeds = \"$SEEDS\"/" ~/.uptickd/config/config.toml
```

:::tip
For more information on seeds and peers, you can the Tendermint [P2P documentation](https://docs.tendermint.com/master/spec/p2p/peer.html).
:::

### Add Persistent Peers

We can set the [`persistent_peers`](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#persistent-peer) field in `~/.uptickd/config/config.toml` to specify peers that your node will maintain persistent connections with. You can retrieve them from the list of
available peers on the [`testnets`](https://github.com/UptickNetwork/uptick-testnet) repo.

```bash
PEERS=`curl -sL https://raw.githubusercontent.com/UptickNetwork/uptick-testnet/main/uptick_7777-1/peers.txt | sort -R | head -n 10 | awk '{print $1}' | paste -s -d, -`
```

Use `sed` to include them into the configuration. You can also add them manually:

```bash
sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" ~/.uptickd/config/config.toml
```
## Start testnet

The final step is to [start the nodes](./../quickstart/run_node#start-node). Once enough voting power (+2/3) from the genesis validators is up-and-running, the testnet will start producing blocks.

```bash
uptickd start
```
## Run a Testnet Validator

Claim your testnet {{ $themeConfig.project.testnet_denom }} on the [faucet](./faucet.md) using your validator account address and submit your validator account address:
> NOTE: Until `uptickd status 2>&1 | jq ."SyncInfo"."catching_up"` got false, create your validator. If your validator is jailed, unjail it via `uptickd tx slashing unjail --from <wallet name> --chain-id uptick_7777-1 -y -b block`.

::: tip
For more details on how to configure your validator, follow the validator [setup](./../guides/validators/setup.md) instructions.
:::
```bash
uptickd tx staking create-validator \
  --amount=5000000000000000000auptick \
  --pubkey=$(uptickd tendermint show-validator) \
  --moniker="UptickBuilder" \
  --chain-id=uptick_7777-1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1000000" \
  --gas="300000" \
  --from=node0 \
  -y \
  -b block
```

## Upgrading Your Node

> NOTE: These instructions are for full nodes that have ran on previous versions of and would like to upgrade to the latest testnet.

### Reset Data

:::warning
If the version <new_version> you are upgrading to is not breaking from the previous one, you **should not** reset the data. If this is the case you can skip to [Restart](#restart)
:::

First, remove the outdated files and reset the data.

```bash
rm $HOME/.uptickd/config/addrbook.json $HOME/.uptickd/config/genesis.json
uptickd unsafe-reset-all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before,
your node will still try to connect to them, but may fail if they haven't also
been upgraded.

::: danger Warning
Make sure that every node has a unique `priv_validator.json`. Do not copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to double sign.
:::

### Restart

To restart your node, just type:

```bash
uptickd start
```
