package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/lib"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Keeper manages connection between chains
type Keeper struct {
	key sdk.StoreKey
	cdc *wire.Codec
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		key: key,
		cdc: cdc,
	}
}

// Channel manages egress queue
type Channel struct {
	k   Keeper
	key sdk.KVStoreGetter
}

func (k Keeper) Channel(key sdk.KVStoreGetter) Channel {
	return Channel{
		k:   k,
		key: key,
	}
}

func (c Channel) egressQueue(chainID string) lib.Linear {
	return lib.NewLinear
}

func (c Channel) Send(ctx sdk.Context, p Payload, dest string, cs sdk.CodespaceType) sdk.Error {
	// TODO: Check validity of the payload; the module have to be permitted to send payload

	packet := Packet{
		Payload:   p,
		SrcChain:  ctx.ChainID(),
		DestChain: dest,
	}

	queue := c.egressQueue(ctx, dest)
	if queue == nil {
		return ErrNoChannelOpened(cs, dest)
	}
	queue.Push(ctx, packet)
	return nil
}
