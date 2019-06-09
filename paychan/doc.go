/*
Package paychan provides unidirectional payment channels.

This module implements simple but feature complete unidirectional payment channels. Channels can be opened by a sender and closed immediately by the receiver, or by the sender subject to a dispute period. There are no top-ups or partial withdrawals (yet). Channels support multiple currencies.


# TODOs
 - Documentation - method descriptions, remove unnecessary comments, docs - Explain how the payment channels are implemented.
 - Refactor submit update route - handle signer detection better

Testing
 - integration test
 	- possible bug in submitting same update repeatedly
 - test cli and rest
 	- verify doesn’t throw json parsing error on invalid json
 	- can’t submit an update from an uninitialised account

Code improvements
 - change channel id to unit64
 - split participants into sender and receiver
 - use iterator for channels for efficiency - rename queue
 - tidy up channel signatures - split off signatures from update as with txs/msgs, can auth sigs be used?
 - custom errors
 - tags
 - clarify naming - paychan vs channel, rename update

Features
 - allow channel signing key to be different from account key
 - configurable channel timeouts
 - use BFT time rather than block height for chanel timeouts
 - sender slashing on early close
 - channel top ups
*/
package paychan
