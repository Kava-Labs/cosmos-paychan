# PayChan - a Cosmos SDK module

[![Go Report Card](https://goreportcard.com/badge/github.com/kava-labs/cosmos-paychan)](https://goreportcard.com/report/github.com/kava-labs/cosmos-paychan)
[![API Reference](https://godoc.org/github.com/kava-labs/cosmos-paychan?status.svg)](https://godoc.org/github.com/kava-labs/cosmos-paychan)
[![License](https://img.shields.io/github/license/kava-labs/cosmos-paychan.svg)](./LICENSE)

*A Cosmos SDK module to add payment channels to any blockchain built using the SDK.*

**Payment Channels** are a technology to speed up payments on blockchains while retaining strong security guarantees. They work by moving payments 'off-chain' to sidestep the bottleneck of blockchain throughput.

**The Cosmos SDK** is a modular framework for developers to quickly and easily build custom blockchains in Go. Blockchains can be built from modules containing specific functionality, such as this one.

> Note: This project is new and unstable. Get involved!


# Usage
This module currently implements unidirectional channels. Channels can be opened by a sender and closed immediately by the receiver, or by the sender subject to a dispute period. There are no top-ups or partial withdrawals (yet).

## 1) Create a channel

	gaiacli tx paychan create cosmos1zls5y0yd9wvh86ceh9tz93eehvesa6d7p8qge8 100atom --from <sender's account name>

## 2) Send off-chain payments
Send a payment for 10 atom.

	gaiacli tx paychan pay <channel ID> 90atom 10atom --filename payment.json

Send the file `payment.json` to your receiver. They can run the following to verify it.

	gaiacli tx paychan close --dry-run --payment payment.json

## 3) Close the channel
The receiver can close immediately at any time.

	gaiacli tx paychan close --from <receiver's account name> --payment payment.json

The sender can submit a close request, closing the channel after a dispute period. During this period a receiver can still close immediately, overruling the sender's request.

	gaiacli tx paychan close --from <sender's account name> --payment payment.json


# Installation
> The aim is for this module to be usable in any cosmos sdk based blockchain. However the module interface in the sdk is currently being refactored so using this module may require some tweaks. See [`go.mod`](./go.mod) for the sdk version this was built against.

This module can be included in an sdk app in the same way the standard modules are (`staking`, `gov`, etc). It uses the module interface pattern introduced in sdk v0.35.0.  
It includes:
 - cli and rest interfaces
 - tx types and handler
 - endblocker to close payment channels

<!--
## User Interfaces
### Rest API
All get request can be used with websockets to subscribe to changes
 - GET  /paychans/{id}
 - GET  /paychans?sender={sAddr}&receiver={rAddr}
 - GET  /paychans/{id}/submitted-update
 - POST /paychans/
 - POST /paychan/{id}/submitted-update (for verifying sigs, use simulate flag in post body)
### Command Line
 - query
   - paychan {id}
   - paychans -sender {sender} -receiver {receiver}
   - submitted-update {id}
 - tx
   - create
   - close
   - pay
-->

# TODOs

#### Features
 - layer 2 utilities - receiver http server and channel watcher
 - configurable channel timeouts
 - use BFT time rather than block height for chanel timeouts
 - allow channel signing key to be different from account key
 - sender slashing on early close
 - channel top ups and partial withdrawals
 - bidirectional channels (with cooperative close)
 - multiparty channels

#### Testing
 - integration test
 	- possible bug in submitting same update repeatedly
 - test cli and rest
 	- verify doesn’t throw json parsing error on invalid json
 	- can’t submit an update from an uninitialised account

#### Code improvements
 - pin to cosmos-sdk v0.36.0 once released
 - change channel id to unit64
 - split participants into sender and receiver
 - use iterator for channels for efficiency - rename queue
 - tidy up channel signatures - split off signatures from update as with txs/msgs, can auth sigs be used?
 - custom errors
 - tags
 - clarify naming - paychan vs channel, rename update