package main

import (
	"fmt"
	"context"
	"math/rand"
	"time"

	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/wire"

	"github.com/matheusd/offline-tx-signer/pkg/offlinetxs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "github.com/decred/dcrwallet/rpc/walletrpc"
)

type destination struct {
	addr dcrutil.Address
	amount dcrutil.Amount
}


type walletClient struct {
	conn        *grpc.ClientConn
	wsvc        pb.WalletServiceClient
	chainParams *chaincfg.Params
}

func connectToWallet(walletHost string, walletCert string, chainParams *chaincfg.Params) (*walletClient, error) {
	rand.Seed(time.Now().Unix())
	creds, err := credentials.NewClientTLSFromFile(walletCert, "localhost")
	if err != nil {
		return nil, err
	}

	optCreds := grpc.WithTransportCredentials(creds)

	connCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	conn, err := grpc.DialContext(connCtx, walletHost, optCreds)
	if err != nil {
		return nil, err
	}

	wsvc := pb.NewWalletServiceClient(conn)

	wc := &walletClient{
		conn:        conn,
		wsvc:        wsvc,
		chainParams: chainParams,
	}

	return wc, nil
}

func (wc *walletClient) genTx(srcAccount uint32, dests []*destination) (*offlinetxs.UnsignedTx, error) {

	ctx := context.TODO()

	addrIndxs := make([]*offlinetxs.AddrIndex, 0)

	outputs := make([]*pb.ConstructTransactionRequest_Output, len(dests))
	for i, dest := range dests {
		outputs[i] = &pb.ConstructTransactionRequest_Output{
			Amount: int64(dest.amount),
			Destination: &pb.ConstructTransactionRequest_OutputDestination{
				Address: dest.addr.EncodeAddress(),
			},
		}

		vaReq := &pb.ValidateAddressRequest{
			Address: dest.addr.EncodeAddress(),
		}
		vaResp, err := wc.wsvc.ValidateAddress(ctx, vaReq)
		if err != nil {
			return nil, err
		}

		if !vaResp.IsMine {
			continue
		}

		branch := uint32(0)
		if vaResp.IsInternal {
			branch = 1
		}

		addrIndxs = append(addrIndxs, &offlinetxs.AddrIndex{
			Account: vaResp.AccountNumber,
			Branch: branch,
			Index: vaResp.Index,
		})
	}

	req := &pb.ConstructTransactionRequest{
		FeePerKb:              0,
		RequiredConfirmations: 0,
		SourceAccount:         srcAccount,
		NonChangeOutputs:      outputs,
	}

	resp, err := wc.wsvc.ConstructTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx()
	err = tx.FromBytes(resp.GetUnsignedTransaction())
	if err != nil {
		return nil, err
	}


	prevTxs := make([]*wire.MsgTx, len(tx.TxIn))
	for i, in := range tx.TxIn{
		inreq := &pb.GetTransactionRequest{
			TransactionHash: in.PreviousOutPoint.Hash[:],
		}

		respIn, err := wc.wsvc.GetTransaction(context.TODO(), inreq)
		if err != nil {
			return nil, err
		}

		prevTxs[i] = wire.NewMsgTx()
		err = prevTxs[i].FromBytes(respIn.GetTransaction().GetTransaction())
		if err != nil {
			return nil, err
		}

		prevOut := prevTxs[i].TxOut[in.PreviousOutPoint.Index]
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(prevOut.Version,
			prevOut.PkScript, wc.chainParams)
		if err != nil {
			return nil, err
		}

		if len(addrs) != 1 {
			return nil, fmt.Errorf("Not a single address")
		}

		vaReq := &pb.ValidateAddressRequest{
			Address: addrs[0].EncodeAddress(),
		}
		vaResp, err := wc.wsvc.ValidateAddress(ctx, vaReq)
		if err != nil {
			return nil, err
		}

		branch := uint32(0)
		if vaResp.IsInternal {
			branch = 1
		}

		addrIndxs = append(addrIndxs, &offlinetxs.AddrIndex{
			Account: vaResp.AccountNumber,
			Branch: branch,
			Index: vaResp.Index,
		})
	}

	return &offlinetxs.UnsignedTx{
		AddrIdxs: addrIndxs,
		PrevTxs: prevTxs,
		Tx: tx,
	}, nil
}
