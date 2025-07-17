# ZK Oracle 
ZK Oracle is a specialized Cosmos client designed specifically for integration with the ZkOS system. Its primary function is to monitor the NYKS blockchain and identify transactions relevant to ZkOS operations.
ZK Oracle leverages the Ignite CLI to query and retrieve transactions from the NYKS chain. It consists of two core components:
1. API Server
The API server provides an interface for users to submit their ZKOS dark transactions directly to the ZkOS core.
2. WebSocket
The WebSocket is utilized by the ZkOS system to receive real-time updates. When ZkOracle detects a transaction of interest on the NYKS chain, it broadcasts the relevant transaction data through this WebSocket channel.

## Installation
### Setup Environment
Follow below steps to setup the oracle
  1. Install [golang](https://go.dev/).
  2. Install [ignite-cli](https://docs.ignite.com/welcome).
  3. Clone the [ZKOracle repository](https://github.com/twilight-project/zkoracle-go.git).

### Run ZK Oracle 
Before running ZK Oracle make sure that you have completely setup the environment, then follow the below step.
update the parameters in the config/config.json file. e.g.
```
{
"accountName": "validator-staging",
"nyks_url": "https://nyks-staging.com/rest/"
}
```
Run below commands on terminal.
```
// install dependencies
go mod tidy
// build
go build .
//run 
go run 
```
It is recommended that you run zkoracle inside a process manager e.g. [supervisord](https://supervisord.org/installing.html) or [systemd](https://systemd.io/)
please refer to their documentation for more details.


## API 

ZK Oracle has 2 API endpoints.

1. Transaction: regular tx to move dark funds around.
```
curl -X POST http://localhost:7000/transaction \
-H "Content-Type: application/json" \
-d '{
  "Txid": "sample-tx-id",
  "Tx": "sample-tx-bytecode",
  "Fee": 1000
}'
```
2. Burnmessage: used to initiate withdrawal of dark fund and minting of lit funds.
```
curl -X POST http://localhost:7000/burnmessage \
-H "Content-Type: application/json" \
-d '{
  "BtcValue": 5000,
  "QqAccount": "sample-qq-account",
  "EncryptScalar": "sample-encrypt-scalar",
  "TwilightAddress": "sample-twilight-address"
}'
```
