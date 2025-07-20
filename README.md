# CometKMS - Remote Key Management Service

> ❗️ **Warning:** FOR DEMO AND TESTING PURPOSES ONLY
> DO NOT USE IN PRODUCTION

Current Diagram:

![CometKMS Architecture](CometKMS.png)

This is under development to be a (high availability) remote signer that uses CometBFT consensus algorithm for the HA layer.
CometBFT decides which node will sign the block requested by an external CometBFT chain and returns the signed block to the chain.
The last signed state is stored as tx in BadgerDB appState as a key-value pair similar to:

```
key: <height>:<round>:<step(type)>
value: {
  "height": <height>,
  "round": <round>,
  "step": <step>,
  "signature": <signature>
}
```
