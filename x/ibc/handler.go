package ibc

import (
	"reflect"

	//"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgOpenConnection:
			return handleMsgOpenConnection(ctx, k, msg)
		case MsgUpdateConnection:
			return handleMsgUpdateConnection(ctx, k, msg)
		default:
			errMsg := "Unrecognized IBC Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgOpenConnection(ctx sdk.Context, k Keeper, msg MsgOpenConnection) sdk.Result {
	store := ctx.KVStore(k.key)
	if k.isConnectionEstablished(store, msg.SrcChain) {
		return ErrConnectionAlreadyEstablished(k.codespace).Result()
	}

	height := msg.ROT.Height()
	if height < 0 {
		return ErrInvalidHeight(k.codespace).Result()
	}
	k.setCommit(store, msg.SrcChain, uint64(msg.ROT.Height()), msg.ROT)

	return sdk.Result{}
}

func handleMsgUpdateConnection(ctx sdk.Context, k Keeper, msg MsgUpdateConnection) sdk.Result {
	store := ctx.KVStore(k.key)
	lastheight, ok := k.getLastCommitHeight(store, msg.SrcChain)
	if !ok {
		return ErrConnectionNotEstablished(k.codespace).Result()
	}

	_ /*lastcommit*/, ok = k.getCommit(store, msg.SrcChain, lastheight)
	if !ok {
		panic("Last commit not found")
	}
	// TODO: add lc verificatioon
	/*
		cert := lite.NewDynamicCertifier(msg.SrcChain, commit.Validators, height)
		if err := cert.Update(msg.Commit); err != nil {
			return ErrUpdateCommitFailed(k.codespace, err).Result()
		}

		k.setCommit(ctx, msg.SrcChain, msg.Commit.Height(), msg.Commit)
	*/
	height := msg.Commit.Commit.Height()
	if height < 0 {
		return ErrInvalidHeight(k.codespace).Result()
	}
	k.setCommit(store, msg.SrcChain, uint64(height), msg.Commit)
	return sdk.Result{}
}

// ----------------------------

func (c Channel) Send(ctx sdk.Context, p Payload, dest string, cs sdk.CodespaceType) sdk.Error {
	// TODO: Check validity of the payload; the module have to be permitted to send payload

	store := c.key.KVStore(ctx)

	packet := Packet{
		Payload:   p,
		SrcChain:  ctx.ChainID(),
		DestChain: dest,
	}

	queue := egressQueue(store, c.k.cdc, dest)
	queue.Push(packet)

	return nil
}

type ReceiveHandler func(sdk.Context, Payload) (Payload, sdk.Error)

func (c Channel) Receive(h ReceiveHandler, ctx sdk.Context, msg MsgReceive) sdk.Result {
	store := ctx.KVStore(c.k.key)
	if err := msg.Verify(store, c); err != nil {
		return err.Result()
	}

	packet := msg.Packet
	if packet.DestChain != ctx.ChainID() {
		return ErrChainMismatch(c.k.codespace).Result()
	}

	cctx, write := ctx.CacheContext()
	receipt, err := h(cctx, packet.Payload)
	if receipt != nil {
		// TODO: check validity of the payload; the handler have to be permitted to send receipt

		packet := Packet{
			Payload:   receipt,
			SrcChain:  ctx.ChainID(),
			DestChain: packet.SrcChain,
		}

		queue := receiptQueue(store, c.k.cdc, packet.SrcChain)
		queue.Push(packet)
	}
	if err != nil {
		return sdk.Result{
			Code: sdk.ABCICodeOK,
			Log:  err.ABCILog(),
		}
	}
	write()

	return sdk.Result{}
}

type ReceiptHandler func(sdk.Context, Payload)

func (c Channel) Receipt(h ReceiptHandler, ctx sdk.Context, msg MsgReceipt) sdk.Result {
	store := ctx.KVStore(c.k.key)
	if err := msg.Verify(store, c); err != nil {
		return err.Result()
	}
	setIngressReceiptSequence(store, c.k.cdc, msg.Packet.SrcChain, msg.Proof.Sequence)

	h(ctx, msg.Payload)

	return sdk.Result{}
}

func handleMsgReceiveCleanup(ctx sdk.Context, c Channel, msg MsgReceiveCleanup) sdk.Result {
	_ = egressQueue(ctx.KVStore(c.k.key), c.k.cdc, msg.SrcChain)
	/*
		if err := msg.Verify(ctx, queue, msg.SrcChain, msg.Sequence); err != nil {
			return err.Result()
		}
	*/
	// TODO: cleanup

	return sdk.Result{}
}

func handleMsgReceiptCleanup(ctx sdk.Context, c Channel, msg MsgReceiptCleanup) sdk.Result {
	_ = receiptQueue(ctx.KVStore(c.k.key), c.k.cdc, msg.SrcChain)
	/*
		if err := msg.Verify(ctx, queue, msg.SrcChain, msg.Sequence); err != nil {
			return err.Result()
		}
	*/
	// TODO: cleanup

	return sdk.Result{}
}
