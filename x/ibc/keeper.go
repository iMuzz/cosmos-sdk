package ibc

import (
	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/lib"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Keeper manages connection between chains
type Keeper struct {
	key sdk.StoreKey
	cdc *wire.Codec

	codespace sdk.CodespaceType
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		key: key,
		cdc: cdc,

		codespace: codespace,
	}
}

// GetLastCommitHeightKey :: []byte -> uint64
func GetLastCommitHeightKey(srcChain []byte) []byte {
	return append([]byte{0x00}, srcChain...)
}

// GetCommitByHeightPrefix :: []byte -> lib.List
func GetCommitByHeightPrefix(srcChain []byte) []byte {
	return append([]byte{0x01}, srcChain...)
}

func commitByHeight(store sdk.KVStore, cdc *wire.Codec, chainID []byte) lib.List {
	return lib.NewList(cdc, store.Prefix(GetCommitByHeightPrefix(chainID)))
}

func (k Keeper) getLastCommitHeight(ctx sdk.Context, srcChain []byte) (res uint64, ok bool) {
	store := ctx.KVStore(k.key)
	bz := store.Get(GetLastCommitHeightKey(srcChain))
	if bz == nil {
		return res, false
	}
	k.cdc.MustUnmarshalBinary(bz, &res)
	return res, true
}

func (k Keeper) getCommit(ctx sdk.Context, srcChain []byte, height uint64) (res lite.FullCommit, ok bool) {
	store := ctx.KVStore(k.key)
	commits := commitByHeight(store, k.cdc, srcChain)
	if err := commits.Get(height, &res); err != nil {
		return res, false
	}
	return res, true
}

func (k Keeper) setCommit(ctx sdk.Context, srcChain []byte, height uint64, commit lite.FullCommit) {
	store := ctx.KVStore(k.key)
	commitByHeight(store, k.cdc, srcChain).Set(height, commit)
	store.Set(GetLastCommitHeightKey(srcChain), k.cdc.MustMarshalBinary(height))
}

func (k Keeper) isConnectionEstablished(ctx sdk.Context, srcChain []byte) bool {
	store := ctx.KVStore(k.key)
	_, ok := k.getLastCommitHeight(ctx, srcChain)
	return ok
}

// Channel manages single channel on a connection
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

// GetEgressQueuePrefix :: lib.Linear
func GetEgressQueuePrefix(destChain []byte) []byte {
	return []byte{0x00}
}

// GetReceiptQueuePrefix :: lib.Linear
func GetReceiptQueuePrefix(destChain []byte) []byte {
	return []byte{0x01}
}

// GetReceivingSequenceKey :: uint64
func GetReceivingSequenceKey(srcChain []byte) []byte {
	return []byte{0x02}
}

func egressQueue(store sdk.KVStore, cdc *wire.Codec, chainID string) lib.Linear {
	return lib.NewLinear(cdc, store.Prefix([]byte{0x00}))
}

func receiptQueue(store sdk.KVStore, cdc *wire.Codec, chainID string) lib.Linear {
	return lib.NewLinear(cdc, store.Prefix([]byte{0x01}))
}

func (c Channel) Send(ctx sdk.Context, p Payload, dest string, cs sdk.CodespaceType) sdk.Error {
	// TODO: Check validity of the payload; the module have to be permitted to send payload

	store := c.key.KVStore(ctx)

	packet := Packet{
		Payload:   p,
		SrcChain:  ctx.ChainID(),
		DestChain: dest,
	}

	queue := egressQueue(store, c.k.cdc, dest)
	if queue == nil {
		return ErrChannelNotOpened(c.k.codespace)
	}
	queue.Push(packet)

	return nil
}
