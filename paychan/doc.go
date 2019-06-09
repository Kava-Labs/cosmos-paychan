/*
Package paychan provides unidirectional payment channels.

This module implements simple but feature complete unidirectional payment channels. Channels can be opened by a sender and closed immediately by the receiver, or by the sender subject to a dispute period. There are no top-ups or partial withdrawals (yet). Channels support multiple currencies.

>Note: This module is still a bit rough around the edges. More feature planned. More test cases needed.


TODO Explain how the payment channels are implemented.

# TODO
 - see if the signature types used by auth can be reused here
 - Documentation - method descriptions, remove unnecessary comments, docs
 - make closeChannel safer?
 - Consider using bft time for channel timeouts.
 - Refactor submit update route - handle signer detection better
 - Implement ValidateBasic for messages
 - Find a better name for Queue - clarify distinction between int slice and abstract queue concept
 - find nicer name for payout
 - write some sort of integration test
 	- possible bug in submitting same update repeatedly
 - add Gas usage
 - add tags (return channel id on creation)
 - refactor cmds to be able to test them, then test them
 	- verify doesn’t throw json parsing error on invalid json
 	- can’t submit an update from an unitialised account
 	- pay without a --from returns confusing error
 - use custom errors instead of using sdk.ErrInternal
 - split off signatures from update as with txs/msgs - testing easier, code easier to use, doesn't store sigs unecessarily on chain
 - consider removing pubKey from UpdateSignature - instead let channel module access accountMapper
 - refactor queue into one object
 - remove printout during tests caused by mock app initialisation
 - with a global channel ID counter, there is a finite number of possible channels ever
 - Should sender close only allow refund of initial amount (not arbitrary payout)? And should a receiver overriding it result in sender slashing?
 - configurable channel 	timeouts

*/
package paychan
