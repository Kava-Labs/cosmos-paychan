package paychan

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/kava-labs/cosmos-paychan/paychan/types"
)

func TestKeeper(t *testing.T) {
	t.Run("CreateChannel", func(t *testing.T) {

		// TODO test for receiver account not existing (OK) and sender not existing (not ok)

		accountSeeds := []string{"senderSeed", "receiverSeed"}
		const (
			senderAccountIndex   int = 0
			receiverAccountIndex int = 1
		)
		_, addrs, _, _ := createTestGenAccounts(accountSeeds, sdk.Coins{}) // pure function

		testCases := []struct {
			name                string
			sender              sdk.AccAddress
			receiver            sdk.AccAddress
			coins               sdk.Coins
			shouldCreateChannel bool
			shouldError         bool
		}{
			{
				"HappyPath",
				addrs[senderAccountIndex],
				addrs[receiverAccountIndex],
				sdk.Coins{sdk.NewInt64Coin("usd", 10)},
				true,
				false,
			},
			{
				"NilAddress",
				sdk.AccAddress{},
				sdk.AccAddress{},
				sdk.Coins{sdk.NewInt64Coin("usd", 10)},
				false,
				true,
			},
			{
				"NilCoins",
				addrs[senderAccountIndex],
				addrs[receiverAccountIndex],
				sdk.Coins{},
				false,
				true,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				////// SETUP
				// create basic mock app
				ctx, coinKeeper, channelKeeper, addrs, _, _, genAccFunding := createMockApp(accountSeeds)
				//
				////// ACTION
				_, err := channelKeeper.CreateChannel(ctx, testCase.sender, testCase.receiver, testCase.coins)

				//
				////// CHECK RESULTS
				// Check error
				if testCase.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				// Check if channel exists and is correct
				channelID := types.ChannelID(0) // should be 0 as first channel
				createdChan, found := channelKeeper.getChannel(ctx, channelID)

				if testCase.shouldCreateChannel {
					expectedChan := types.Channel{
						ID:           channelID,
						Participants: [2]sdk.AccAddress{testCase.sender, testCase.receiver},
						Coins:        testCase.coins,
					}

					// channel exists and correct
					assert.True(t, found)
					assert.Equal(t, expectedChan, createdChan)
					// check coins deducted from sender
					assert.Equal(t, genAccFunding.Sub(testCase.coins), coinKeeper.GetCoins(ctx, testCase.sender))
					// check no coins deducted from receiver
					assert.Equal(t, genAccFunding, coinKeeper.GetCoins(ctx, testCase.receiver))
					// check next global channelID incremented
					assert.Equal(t, types.ChannelID(1), channelKeeper.getNewChannelID(ctx))
				} else {
					// channel doesn't exist
					assert.False(t, found)
					assert.Equal(t, types.Channel{}, createdChan)
					// check no coins deducted from sender
					assert.Equal(t, genAccFunding, coinKeeper.GetCoins(ctx, addrs[senderAccountIndex]))
					// check no coins deducted from receiver
					assert.Equal(t, genAccFunding, coinKeeper.GetCoins(ctx, addrs[receiverAccountIndex]))
					// check next global channelID not incremented
					assert.Equal(t, types.ChannelID(0), channelKeeper.getNewChannelID(ctx))
				}
			})
		}
	})

	t.Run("CloseChannelByReceiver", func(t *testing.T) {
		// TODO convert to table driven and add more test cases
		//		channel exists or not (assume channels correct)
		//		various Updates
		//		submittedUpdates existing or not (assume they are valid)

		// SETUP
		accountSeeds := []string{"senderSeed", "receiverSeed"}
		const (
			senderAccountIndex   int = 0
			receiverAccountIndex int = 1
		)
		ctx, coinKeeper, channelKeeper, addrs, pubKeys, privKeys, genAccFunding := createMockApp(accountSeeds)

		coins := sdk.Coins{sdk.NewInt64Coin("usd", 10)}

		// create new channel
		channelID := types.ChannelID(0) // should be 0 as first channel
		channel := types.Channel{
			ID:           channelID,
			Participants: [2]sdk.AccAddress{addrs[senderAccountIndex], addrs[receiverAccountIndex]},
			Coins:        coins,
		}
		channelKeeper.setChannel(ctx, channel)

		// create closing update
		payout := types.Payout{sdk.Coins{sdk.NewInt64Coin("usd", 3)}, sdk.Coins{sdk.NewInt64Coin("usd", 7)}}
		update := types.Update{
			ChannelID: channelID,
			Payout:    payout,
			// empty sig
		}
		cryptoSig, _ := privKeys[senderAccountIndex].Sign(update.GetSignBytes())
		update.Sigs = [1]types.UpdateSignature{types.UpdateSignature{
			PubKey:          pubKeys[senderAccountIndex],
			CryptoSignature: cryptoSig,
		}}

		// Set empty submittedUpdatesQueue TODO work out proper genesis initialisation
		channelKeeper.setSubmittedUpdatesQueue(ctx, types.SubmittedUpdatesQueue{})

		// ACTION
		_, err := channelKeeper.CloseChannelByReceiver(ctx, update)

		// CHECK RESULTS
		// no error
		assert.NoError(t, err)
		// coins paid out
		senderPayout := payout[senderAccountIndex]
		assert.Equal(t, genAccFunding.Add(senderPayout), coinKeeper.GetCoins(ctx, addrs[senderAccountIndex]))
		receiverPayout := payout[receiverAccountIndex]
		assert.Equal(t, genAccFunding.Add(receiverPayout), coinKeeper.GetCoins(ctx, addrs[receiverAccountIndex]))
		// channel deleted
		_, found := channelKeeper.getChannel(ctx, channelID)
		assert.False(t, found)

	})

	t.Run("InitCloseChannelBySender", func(t *testing.T) {

		// TODO do some documentation here
		// Ideally this should mock calls to ctx.store.Get/Set - test the side effects without being dependent on implementation details
		// TODO test correct behavior when a submittedUpdate already exists

		accountSeeds := []string{"senderSeed", "receiverSeed", "notInChannelSeed"}
		const (
			senderAccountIndex   int = 0
			receiverAccountIndex int = 1
			otherAccountIndex    int = 2
		)
		chanID := types.ChannelID(0)

		type testUpdate struct { // A parameterised version of an Update for use in specifying test cases.
			channelID          types.ChannelID // channelID of submitted update
			payout             types.Payout    // payout of submitted update
			pubKeyAccountIndex int             // pubkey of signature of submitted update
			sigAccountIndex    int             // crypto signature of signature of submitted update
		}
		testCases := []struct {
			name                    string
			setupChannel            bool
			updateToSubmit          testUpdate
			expectedSubmittedUpdate string // "empty" or "sameAsSubmitted"
			shouldError             bool
		}{
			{
				"HappyPath",
				true,
				testUpdate{chanID, types.Payout{sdk.Coins{sdk.NewInt64Coin("usd", 3)}, sdk.Coins{sdk.NewInt64Coin("usd", 7)}}, senderAccountIndex, senderAccountIndex},
				"sameAsSubmited",
				false,
			},
			{
				"NoChannel",
				false,
				testUpdate{chanID, types.Payout{sdk.Coins{sdk.NewInt64Coin("usd", 3)}, sdk.Coins{sdk.NewInt64Coin("usd", 7)}}, senderAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"NoCoins",
				true,
				testUpdate{chanID, types.Payout{sdk.Coins{}}, senderAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"TooManyCoins",
				true,
				testUpdate{chanID, types.Payout{sdk.Coins{sdk.NewInt64Coin("usd", 100)}, sdk.Coins{sdk.NewInt64Coin("usd", 7)}}, senderAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"WrongSignature",
				true,
				testUpdate{chanID, types.Payout{sdk.Coins{sdk.NewInt64Coin("usd", 3)}, sdk.Coins{sdk.NewInt64Coin("usd", 7)}}, senderAccountIndex, otherAccountIndex},
				"empty",
				true,
			},
			{
				"WrongPubKey",
				true,
				testUpdate{chanID, types.Payout{sdk.Coins{sdk.NewInt64Coin("usd", 3)}, sdk.Coins{sdk.NewInt64Coin("usd", 7)}}, otherAccountIndex, senderAccountIndex},
				"empty",
				true,
			},
			{
				"ReceiverSigned",
				true,
				testUpdate{chanID, types.Payout{sdk.Coins{sdk.NewInt64Coin("usd", 3)}, sdk.Coins{sdk.NewInt64Coin("usd", 7)}}, receiverAccountIndex, receiverAccountIndex},
				"empty",
				true,
			},
		}
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {

				// SETUP
				ctx, _, channelKeeper, addrs, pubKeys, privKeys, _ := createMockApp(accountSeeds)
				// Set empty submittedUpdatesQueue TODO work out proper genesis initialisation
				channelKeeper.setSubmittedUpdatesQueue(ctx, types.SubmittedUpdatesQueue{})
				// create new channel
				if testCase.setupChannel {
					channel := types.Channel{
						ID:           chanID, // should be 0 as first channel
						Participants: [2]sdk.AccAddress{addrs[senderAccountIndex], addrs[receiverAccountIndex]},
						Coins:        sdk.Coins{sdk.NewInt64Coin("usd", 10)},
					}
					channelKeeper.setChannel(ctx, channel)
				}

				// create update
				// basic values
				updateToSubmit := types.Update{
					ChannelID: testCase.updateToSubmit.channelID,
					Payout:    testCase.updateToSubmit.payout,
					// empty sig
				}
				// create update's signature
				cryptoSig, _ := privKeys[testCase.updateToSubmit.sigAccountIndex].Sign(updateToSubmit.GetSignBytes())
				updateToSubmit.Sigs = [1]types.UpdateSignature{types.UpdateSignature{
					PubKey:          pubKeys[testCase.updateToSubmit.pubKeyAccountIndex],
					CryptoSignature: cryptoSig,
				}}

				// ACTION
				_, err := channelKeeper.InitCloseChannelBySender(ctx, updateToSubmit)

				// CHECK RESULTS
				// Check error
				if testCase.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				// Check submittedUpdate
				su, found := channelKeeper.getSubmittedUpdate(ctx, chanID)
				switch testCase.expectedSubmittedUpdate {
				case "empty":
					assert.False(t, found)
					assert.Zero(t, su)
				case "sameAsSubmitted":
					assert.True(t, found)
					expectedSU := types.SubmittedUpdate{updateToSubmit, types.ChannelDisputeTime}
					assert.Equal(t, expectedSU, su)
				}

			})
		}

	})

}
