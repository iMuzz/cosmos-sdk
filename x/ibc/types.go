package ibc

import (
	"encoding/json"

	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: lightclient verification

// ---------------------------------
// ReceiveMsg

// ReceiveMsg defines the message that a relayer uses to post a packet
// to the destination chain.

type ReceiveMsg struct {
	Packet
	Proof
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

func (msg ReceiveMsg) Verify(store sdk.KVStore, c Channel) sdk.Error {
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
// ReceiptMsg

type ReceiptMsg struct {
	Packet
	Proof
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

func (msg ReceiptMsg) Verify(store sdk.KVStore, c Channel) sdk.Error {
	chainID := msg.Packet.SrcChain

	expected := getReceiptSequence(store, c.k.cdc, chainID)
	proof := msg.Proof
	if proof.Sequence != uint64(expected) {
		return ErrInvalidSequence(c.k.codespace)
	}

	return proof.Verify(store, msg.Packet)
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
// OpenConnectionMsg

// OpenConnectionMsg defines the message that is used for open a c
// that receives msg from another chain
type OpenConnectionMsg struct {
	ROT      lite.FullCommit
	SrcChain []byte
	Signer   sdk.Address
}

func (msg OpenConnectionMsg) Type() string {
	return "ibc"
}

func (msg OpenConnectionMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg OpenConnectionMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg OpenConnectionMsg) ValidateBasic() sdk.Error {
	return nil
}

func (msg OpenConnectionMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer}
}

//------------------------------------
// UpdateConnectionMsg

type UpdateConnectionMsg struct {
	SrcChain []byte
	Commit   lite.FullCommit
	//PacketProof
	Signer sdk.Address
}

func (msg UpdateConnectionMsg) Type() string {
	return "ibc"
}

func (msg UpdateConnectionMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg UpdateConnectionMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg UpdateConnectionMsg) ValidateBasic() sdk.Error {
	return nil
}

func (msg UpdateConnectionMsg) GetSigners() []sdk.Address {
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
