package offlinetxs

import (
	"fmt"
	"bytes"

	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainec"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/dcrec"
	"github.com/decred/dcrd/hdkeychain"
	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/wire"
)

type accountBranchKeys struct {
	accountKey *hdkeychain.ExtendedKey
	branchKeys map[uint32]*hdkeychain.ExtendedKey
}

type Signer struct {
	addresses map[string]*hdkeychain.ExtendedKey
	chainParams *chaincfg.Params
}

func NewSigner(addrIndexes []*AddrIndex, seed []byte,
	chainParams *chaincfg.Params) (*Signer, error) {

	purpose := uint32(44)
	coinType := chainParams.SLIP0044CoinType

	rootKey, err := hdkeychain.NewMaster(seed, chainParams)
	if err != nil {
		return nil, err
	}

	purposeKey, err := rootKey.Child(purpose + hdkeychain.HardenedKeyStart)
	if err != nil {
		return nil, err
	}

	coinTypeKey, err := purposeKey.Child(coinType + hdkeychain.HardenedKeyStart)
	if err != nil {
		return nil, err
	}

	accountKeys := make(map[uint32]*accountBranchKeys)
	keys := make(map[AddrIndex]*hdkeychain.ExtendedKey)
	addresses := make(map[string]*hdkeychain.ExtendedKey)

	// derive keys for addresses

	for _, idx := range addrIndexes {
		if _, has := keys[*idx]; has {
			continue
		}

		ak, hasAk := accountKeys[idx.Account]
		if !hasAk {
			account, err := coinTypeKey.Child(idx.Account + hdkeychain.HardenedKeyStart)
			if err != nil {
				return nil, err
			}

			ak = &accountBranchKeys{
				accountKey: account,
				branchKeys: make(map[uint32]*hdkeychain.ExtendedKey),
			}
			accountKeys[idx.Account] = ak
		}

		bk, hasBk := ak.branchKeys[idx.Branch]
		if !hasBk {
			bk, err = ak.accountKey.Child(idx.Branch)
			if err != nil {
				return nil, err
			}

			ak.branchKeys[idx.Branch] = bk
		}

		ik, err := bk.Child(idx.Index)
		if err != nil {
			return nil, err
		}
		keys[*idx] = ik

		addr, err := ik.Address(chainParams)
		if err != nil {
			return nil, err
		}
		addresses[addr.EncodeAddress()] = ik
	}

	return &Signer{
		addresses: addresses,
		chainParams: chainParams,
	}, nil
}


func (s *Signer) getKey(addr dcrutil.Address) (chainec.PrivateKey, bool, error) {
	key, has := s.addresses[addr.EncodeAddress()]
	if !has {
		return nil, false, fmt.Errorf("does not have key for address %s", addr.EncodeAddress())
	}

	pk, err := key.ECPrivKey()
	if err != nil {
		return nil, false, err
	}

	pubk, err := key.ECPubKey()
	if err != nil {
		return nil, false, err
	}
	pkhCompressed := dcrutil.Hash160(pubk.SerializeCompressed())

	origHash := addr.Hash160()
	var compressed bool
	if bytes.Equal(pkhCompressed, origHash[:]) {
		compressed = true
	}

	return pk, compressed, nil
}

func (s *Signer) getScript(dcrutil.Address) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Signer) Sign(tx UnsignedTx) (*wire.MsgTx, error) {

	prevTxs := make(map[chainhash.Hash]*wire.MsgTx, len(tx.PrevTxs))
	for _, tx := range tx.PrevTxs {
		prevTxs[tx.TxHash()] = tx
	}

	getKey := txscript.KeyClosure(s.getKey)
	getScript := txscript.ScriptClosure(s.getScript)

	signed := tx.Tx.Copy()
	for i, in := range tx.Tx.TxIn {
		prevTx, has := prevTxs[in.PreviousOutPoint.Hash]
		if !has {
			return nil, fmt.Errorf("tx not found %s", in.PreviousOutPoint.Hash.String())
		}

		prevOut := prevTx.TxOut[in.PreviousOutPoint.Index]
		prevOutScript := prevOut.PkScript

		sigScript, err := txscript.SignTxOutput(s.chainParams,
			tx.Tx, i, prevOutScript, txscript.SigHashAll, getKey,
			getScript, nil, dcrec.STEcdsaSecp256k1)
		if err != nil {
			return nil, err
		}

		signed.TxIn[i].SignatureScript = sigScript
	}

	return signed, nil
}
