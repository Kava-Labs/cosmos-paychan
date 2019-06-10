/*
Package paychan provides unidirectional payment channels for cosmos SDK based blockchains.

This module implements simple but feature complete unidirectional payment channels.
Channels can be opened by a sender and closed immediately by the receiver, or by the sender subject to a dispute period.
There are no top-ups or partial withdrawals (yet). Channels support multiple currencies.
*/
package paychan
