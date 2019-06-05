module github.com/kava-labs/cosmos-sdk-paychan

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.28.2-0.20190603133151-59ac1480617c
	github.com/gorilla/mux v1.7.2
	github.com/kava-labs/kava v0.0.0-20181008134728-118e18441e5c
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.6
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
