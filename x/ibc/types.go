package ibc

import (
	"encoding/json"

	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/lib"
)

// TODO: lightclient verification

// ---------------------------------
// ReceiveMsg

// ReceiveMsg defines the message that a relayer uses to post a packet
// to the destination chain.

type ReceiveMsg struct {
	Packet
	Relayer sdk.Address
}

func (msg ReceiveMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg ReceiveMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg ReceiveMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}

func (msg ReceiveMsg) Verify(ctx sdk.Context, channel Channel) sdk.Error {
	chainID := msg.Packet.SrcChain

	expected := channel.getReceiveSequence(ctx, chainID)
	// TODO: unify int64/uint64
	if proof.Sequence != uint64(expected) {
		return ErrInvalidSequence(channel.keeper.codespace)
	}
	channel.setReceiveSequence(ctx, chainID, int64(proof.Sequence+1))

	return proof.Verify(ctx, channel, msg.Packet)
}

// --------------------------------
// ReceiptMsg

type ReceiptMsg struct {
	Packet
	Relayer sdk.Address
}

func (msg ReceiptMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg ReceiptMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg ReceiptMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}

func (msg ReceiptMsg) Verify(ctx sdk.Context, channel Channel) sdk.Error {
	chainID := msg.Packet.SrcChain

	expected := channel.getReceiptSequence(ctx, chainID)
	if proof.Sequence != uint64(expected) {
		return ErrInvalidSequence(channel.keeper.codespace)
	}
	channel.setReceiptSequence(ctx, chainID, int64(proof.Sequence+1))

	return proof.Verify(ctx, channel, msg.Packet)
}

// --------------------------------
// ReceiveCleanupMsg

type ReceiveCleanupMsg struct {
	ChannelName string
	Sequence    int64
	SrcChain    string
	Cleaner     sdk.Address
}

func (msg ReceiveCleanupMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg ReceiveCleanupMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg ReceiveCleanupMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Cleaner}
}

func (msg ReceiveCleanupMsg) Type() string {
	return "ibc"
}

func (msg ReceiveCleanupMsg) ValidateBasic() sdk.Error {
	return nil
}

// --------------------------------
// ReceiptCleanupMsg

type ReceiptCleanupMsg struct {
	ChannelName string
	Sequence    int64
	SrcChain    string
	Cleaner     sdk.Address
}

func (msg ReceiptCleanupMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg ReceiptCleanupMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg ReceiptCleanupMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Cleaner}
}

func (msg ReceiptCleanupMsg) Type() string {
	return "ibc"
}

func (msg ReceiptCleanupMsg) ValidateBasic() sdk.Error {
	return nil
}

//-------------------------------------
// OpenChannelMsg

// OpenChannelMsg defines the message that is used for open a channel
// that receives msg from another chain
type OpenChannelMsg struct {
	ROT      lite.FullCommit
	SrcChain string
	Signer   sdk.Address
}

func (msg OpenChannelMsg) Type() string {
	return "ibc"
}

func (msg OpenChannelMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg OpenChannelMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg OpenChannelMsg) ValidateBasic() sdk.Error {
	return nil
}

func (msg OpenChannelMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer}
}

//------------------------------------
// UpdateChannelMsg

type UpdateChannelMsg struct {
	SrcChain string
	Commit   lite.FullCommit
	//PacketProof
	Signer sdk.Address
}

func (msg UpdateChannelMsg) Type() string {
	return "ibc"
}

func (msg UpdateChannelMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg UpdateChannelMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg UpdateChannelMsg) ValidateBasic() sdk.Error {
	return nil
}

func (msg UpdateChannelMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer}
}

// ------------------------------
// Payload
// Payload defines inter-blockchain message
// that can be proved by light-client protocol

type Payload interface {
	Type() string
	ValidateBasic() sdk.Error
}

// ------------------------------
// Packet

// Packet defines a piece of data that can be send between two separate
// blockchains.
type Packet struct {
	Payload
	SrcChain  string
	DestChain string
}
