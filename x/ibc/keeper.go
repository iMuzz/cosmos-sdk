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
	return lib.NewList(cdc, store.Prefix(GetCommitByHeightPrefix(chainID)), nil)
}

func (k Keeper) getLastCommitHeight(store sdk.KVStore, srcChain []byte) (res uint64, ok bool) {
	bz := store.Get(GetLastCommitHeightKey(srcChain))
	if bz == nil {
		return res, false
	}
	k.cdc.MustUnmarshalBinary(bz, &res)
	return res, true
}

func (k Keeper) getCommit(store sdk.KVStore, srcChain []byte, height uint64) (res lite.FullCommit, ok bool) {
	commits := commitByHeight(store, k.cdc, srcChain)
	if err := commits.Get(height, &res); err != nil {
		return res, false
	}
	return res, true
}

func (k Keeper) setCommit(store sdk.KVStore, srcChain []byte, height uint64, commit lite.FullCommit) {
	commitByHeight(store, k.cdc, srcChain).Set(height, commit)
	store.Set(GetLastCommitHeightKey(srcChain), k.cdc.MustMarshalBinary(height))
}

func (k Keeper) isConnectionEstablished(store sdk.KVStore, srcChain []byte) bool {
	_, ok := k.getLastCommitHeight(store, srcChain)
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

// GetEgressQueuePrefix :: string -> lib.Linear
func GetEgressQueuePrefix(destChain string) []byte {
	return append([]byte{0x00}, []byte(destChain)...)
}

// GetEgressReceiptQueuePrefix :: string -> lib.Linear
func GetEgressReceiptQueuePrefix(destChain string) []byte {
	return append([]byte{0x01}, []byte(destChain)...)
}

// GetIngressSequenceKey :: string -> uint64
func GetIngressSequenceKey(srcChain string) []byte {
	return append([]byte{0x02}, []byte(srcChain)...)
}

// GetIngressReceiptSequenceKey :: uint64
func GetIngressReceiptSequenceKey(srcChain string) []byte {
	return append([]byte{0x03}, []byte(srcChain)...)
}

func egressQueue(store sdk.KVStore, cdc *wire.Codec, chainID string) lib.Linear {
	return lib.NewLinear(cdc, store.Prefix([]byte{0x00}), nil)
}

func receiptQueue(store sdk.KVStore, cdc *wire.Codec, chainID string) lib.Linear {
	return lib.NewLinear(cdc, store.Prefix([]byte{0x01}), nil)
}

func getIngressSequence(store sdk.KVStore, cdc *wire.Codec, srcChain string) (res uint64) {
	bz := store.Get(GetIngressSequenceKey(srcChain))
	if bz == nil {
		return 0
	}
	cdc.MustUnmarshalBinary(bz, &res)
	return
}

func setRecivingSequence(store sdk.KVStore, cdc *wire.Codec, srcChain string, seq uint64) {
	store.Set(GetIngressSequenceKey(srcChain), cdc.MustMarshalBinary(seq))
}

func getIngressReceiptSequence(store sdk.KVStore, cdc *wire.Codec, srcChain string) (res uint64) {
	bz := store.Get(GetIngressReceiptSequenceKey(srcChain))
	if bz == nil {
		return 0
	}
	cdc.MustUnmarshalBinary(bz, &res)
	return
}

func setIngressReceiptSequence(store sdk.KVStore, cdc *wire.Codec, srcChain string, seq uint64) {
	store.Set(GetIngressReceiptSequenceKey(srcChain), cdc.MustMarshalBinary(seq))
}
