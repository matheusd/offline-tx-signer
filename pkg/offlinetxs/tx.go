package offlinetxs

import (
	"encoding/base64"
	"encoding/json"

	"github.com/decred/dcrd/wire"
	"github.com/mitchellh/mapstructure"
)

type AddrIndex struct {
	Account uint32 `json:"account"`
	Branch  uint32 `json:"branch"`
	Index   uint32 `json:"index"`
}

type UnsignedTx struct {
	Tx       *wire.MsgTx
	PrevTxs  []*wire.MsgTx
	AddrIdxs []*AddrIndex
}

func (tx *UnsignedTx) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})

	txBytes, err := tx.Tx.Bytes()
	if err != nil {
		return nil, err
	}
	prevTxsBytes := make([][]byte, len(tx.PrevTxs))
	for i, prev := range tx.PrevTxs {
		b, err := prev.Bytes()
		if err != nil {
			return nil, err
		}
		prevTxsBytes[i] = b
	}

	m["addrIdxs"] = tx.AddrIdxs
	m["tx"] = txBytes
	m["prevTxs"] = prevTxsBytes

	return json.Marshal(m)
}

func (tx *UnsignedTx) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	b64enc := base64.StdEncoding
	txBts, err := b64enc.DecodeString(m["tx"].(string))
	if err != nil {
		return err
	}

	tx.Tx = wire.NewMsgTx()
	err = tx.Tx.FromBytes(txBts)
	if err != nil {
		return err
	}

	prevTxs := m["prevTxs"].([]interface{})
	tx.PrevTxs = make([]*wire.MsgTx, len(prevTxs))
	for i, intf := range prevTxs {
		prevTxBts, err := b64enc.DecodeString(intf.(string))
		if err != nil {
			return err
		}
		tx.PrevTxs[i] = wire.NewMsgTx()
		err = tx.PrevTxs[i].FromBytes(prevTxBts)
		if err != nil {
			return nil
		}
	}

	err = mapstructure.Decode(m["addrIdxs"], &tx.AddrIdxs)
	if err != nil {
		return err
	}

	return nil
}
