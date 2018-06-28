package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// IBC transfer command
func IBCSendTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use: "transfer",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// build the message
			msg, err := buildMsg(from)
			if err != nil {
				return err
			}

			// get password
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().String(FlagTo, "", "Address to send coins")
	cmd.Flags().String(FlagAmount, "", "Amount of coins to send")
	cmd.Flags().String(FlagDestChain, "", "Destination chain to send coins")
	return cmd
}

func buildMsg(from sdk.Address) (sdk.Msg, error) {
	amount := viper.GetString(FlagAmount)
	coins, err := sdk.ParseCoins(amount)
	if err != nil {
		return nil, err
	}

	dest := viper.GetString(FlagTo)
	bz, err := hex.DecodeString(dest)
	if err != nil {
		return nil, err
	}
	to := sdk.Address(bz)

	payload := bank.PayloadSend{
		SrcAddr:  from,
		DestAddr: to,
		Coins:    coins,
	}

	msg := bank.MsgIBCSend{
		PayloadSend: payload,
		DestChain:   viper.GetString(FlagDestChain),
	}

	return msg, nil
}
