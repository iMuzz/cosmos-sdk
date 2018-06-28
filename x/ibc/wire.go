package ibc

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgReceive{}, "cosmos-sdk/Receive", nil)
	cdc.RegisterConcrete(MsgReceipt{}, "cosmos-sdk/Receipt", nil)
	cdc.RegisterConcrete(MsgReceiveCleanup{}, "cosmos-sdk/ReceiveCleanup", nil)
	cdc.RegisterConcrete(MsgReceiptCleanup{}, "cosmos-sdk/ReceiptCleanup", nil)
	cdc.RegisterConcrete(MsgOpenConnection{}, "cosmos-sdk/OpenConnection", nil)
	cdc.RegisterConcrete(MsgUpdateConnection{}, "cosmos-sdk/UpdateConnection", nil)

	cdc.RegisterInterface((*Payload)(nil), nil)
}
