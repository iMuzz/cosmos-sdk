package ibc

import (
	"encoding/json"

	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: lightclient verification

// ---------------------------------
// MsgReceive

// MsgReceive defines the message that a relayer uses to post a packet
// to the destination chain.

type MsgReceive struct {
	Packet
	Proof
	Relayer sdk.Address
}

func (msg MsgReceive) Get(key interface{}) interface{} {
	return nil
}

func (msg MsgReceive) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgReceive) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}

func (msg MsgReceive) Verify(store sdk.KVStore, c Channel) sdk.Error {
	chainID := msg.Packet.SrcChain

	expected := egressQueue(store, c.k.cdc, chainID)
	// TODO: unify int64/uint64
	proof := msg.Proof
	if proof.Sequence != uint64(expected.Len()) {
		return ErrInvalidSequence(c.k.codespace)
	}

	return proof.Verify(store, msg.Packet)
}

// --------------------------------
// MsgReceipt

type MsgReceipt struct {
	Packet
	Proof
	Relayer sdk.Address
}

func (msg MsgReceipt) Get(key interface{}) interface{} {
	return nil
}

func (msg MsgReceipt) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgReceipt) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}

func (msg MsgReceipt) Verify(store sdk.KVStore, c Channel) sdk.Error {
	chainID := msg.Packet.SrcChain

	expected := getIngressReceiptSequence(store, c.k.cdc, chainID)
	proof := msg.Proof
	if proof.Sequence != uint64(expected) {
		return ErrInvalidSequence(c.k.codespace)
	}

	return proof.Verify(store, msg.Packet)
}

// --------------------------------
// MsgReceiveCleanup

type MsgReceiveCleanup struct {
	ChannelName string
	Sequence    int64
	SrcChain    string
	Cleaner     sdk.Address
}

func (msg MsgReceiveCleanup) Get(key interface{}) interface{} {
	return nil
}

func (msg MsgReceiveCleanup) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgReceiveCleanup) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Cleaner}
}

func (msg MsgReceiveCleanup) Type() string {
	return "ibc"
}

func (msg MsgReceiveCleanup) ValidateBasic() sdk.Error {
	return nil
}

// --------------------------------
// MsgReceiptCleanup

type MsgReceiptCleanup struct {
	ChannelName string
	Sequence    int64
	SrcChain    string
	Cleaner     sdk.Address
}

func (msg MsgReceiptCleanup) Get(key interface{}) interface{} {
	return nil
}

func (msg MsgReceiptCleanup) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgReceiptCleanup) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Cleaner}
}

func (msg MsgReceiptCleanup) Type() string {
	return "ibc"
}

func (msg MsgReceiptCleanup) ValidateBasic() sdk.Error {
	return nil
}

//-------------------------------------
// MsgOpenConnection

// MsgOpenConnection defines the message that is used for open a c
// that receives msg from another chain
type MsgOpenConnection struct {
	ROT      lite.FullCommit
	SrcChain []byte
	Signer   sdk.Address
}

func (msg MsgOpenConnection) Type() string {
	return "ibc"
}

func (msg MsgOpenConnection) Get(key interface{}) interface{} {
	return nil
}

func (msg MsgOpenConnection) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgOpenConnection) ValidateBasic() sdk.Error {
	return nil
}

func (msg MsgOpenConnection) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer}
}

//------------------------------------
// MsgUpdateConnection

type MsgUpdateConnection struct {
	SrcChain []byte
	Commit   lite.FullCommit
	//PacketProof
	Signer sdk.Address
}

func (msg MsgUpdateConnection) Type() string {
	return "ibc"
}

func (msg MsgUpdateConnection) Get(key interface{}) interface{} {
	return nil
}

func (msg MsgUpdateConnection) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgUpdateConnection) ValidateBasic() sdk.Error {
	return nil
}

func (msg MsgUpdateConnection) GetSigners() []sdk.Address {
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

// ------------------------------
// Proof

type Proof struct {
	// Proof merkle.Proof
	Height   uint64
	Sequence uint64
}

func (prf Proof) Verify(store sdk.KVStore, p Packet) sdk.Error {
	// TODO: implement
	return nil
}
