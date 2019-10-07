package handlers

import (
	"context"
	"net/http"
	"encoding/json"
	"fmt"
	"github.com/coschain/contentos-go/prototype"
	"github.com/coschain/contentos-go/rpc"
	"github.com/coschain/contentos-go/rpc/pb"
	"encoding/binary"
	"hash/crc32"
	"errors"
	"strconv"
)

type VoteReq struct {
	AccountName string `json:"account_name"`
	PrivateKey string `json:"private_key"`
	PostId string `json:"post_id"`
}

type VoteRes struct {
	Status int `json:"status"`
	Message string `json:"message"`
}

// var conn, _ = rpc.Dial("34.203.85.235:8888")
var conn, _ = rpc.Dial("34.207.44.234:8888")
var rpcClient = grpcpb.NewApiServiceClient(conn)
// var mChainIdName string = "dev"
var mChainIdName = "main"


func VotePost(write http.ResponseWriter, request *http.Request) {
	var voteReq VoteReq
	err := json.NewDecoder(request.Body).Decode(&voteReq)

	if err != nil {
		http.Error(write, err.Error(), 400)
		return
	}

	postIdInt64, err := strconv.ParseUint(voteReq.PostId, 10, 64)

	if err != nil {
		http.Error(write, err.Error(), 500)
		return
	}

	errVote := vote(voteReq.AccountName, voteReq.PrivateKey, postIdInt64)

	var voteRes VoteRes

	voteRes.Status = 1
	voteRes.Message = "Success"

	if errVote != nil {
		voteRes.Status = 2
		voteRes.Message = errVote.Error()
	}

	output, err := json.Marshal(voteRes)

	if err != nil {
		http.Error(write, err.Error(), 500)
		return
	}

	write.Header().Set("Content-Type", "application/json")
	write.Write(output)
}

func vote(accountNem string, privateKey string, postId uint64) error {
	voteOP := &prototype.VoteOperation {
		Voter: prototype.NewAccountName(accountNem),
		Idx:   postId,
	}

	signedTx, err := signTx(privateKey, voteOP)

	if err != nil {
		fmt.Println("signedTxerr", err.Error())
		return err
	}

	errReq := txRequest(signedTx)

	if errReq != nil {
		fmt.Println("txRequesterr", errReq.Error())
		return errReq
	}

	return nil
}

func signTx(privateKey string, ops ...interface{}) (*prototype.SignedTransaction, error) {
	privKey := &prototype.PrivateKeyType{}
	pk, err := prototype.PrivateKeyFromWIF(privateKey)

	if err != nil {
		fmt.Println("PrivateKeyFromWIFerr", err.Error())
		return nil, err
	}
	privKey = pk

	chainState, err := getChainState()

	if err != nil {
		fmt.Println("getChainStateerr", err.Error())
		return nil, err
	}

	refBlockPrefix := binary.BigEndian.Uint32(chainState.Dgpo.HeadBlockId.Hash[8:12])
	refBlockNum := uint32(chainState.Dgpo.HeadBlockNumber & 0x7ff)
	tx := &prototype.Transaction{RefBlockNum: refBlockNum, RefBlockPrefix: refBlockPrefix, Expiration: &prototype.TimePointSec{UtcSeconds: chainState.Dgpo.Time.UtcSeconds + 55}}

	for _, op := range ops {
		tx.AddOperation(op)
	}

	signTx := prototype.SignedTransaction{Trx: tx}
	chainIdName := mChainIdName
	chainId := crc32.ChecksumIEEE([]byte(chainIdName))

	res := signTx.Sign(privKey, prototype.ChainId{Value: chainId})
	signTx.Signature = &prototype.SignatureType{Sig: res}

	if err := signTx.Validate(); err != nil {
		fmt.Println("signTx.Validateerr", err.Error())
		return nil, err
	}

	return &signTx, nil
}

func getChainState() (*grpcpb.ChainState, error) {
	req := &grpcpb.NonParamsRequest{}
	res, err := rpcClient.GetChainState(context.Background(), req)

	if err != nil {
		fmt.Println("getChainStateerr", err.Error())
		return nil, err
	}

	if res == nil {
		fmt.Println("getChainStateresnil")
		return nil, errors.New("getChainState res: nil")
	}
	return res.State, nil
}

func txRequest(signedTx *prototype.SignedTransaction) error {
	req := &grpcpb.BroadcastTrxRequest{Transaction: signedTx, Finality: false}
	res, err := rpcClient.BroadcastTrx(context.Background(), req)

	if err != nil {
		fmt.Println("BroadcastTrxerr", err.Error())
		return err
	}

	if res == nil {
		fmt.Println("BroadcastTrx res nil")
		return errors.New("BroadcastTrx res nil")
	}

	if res.Invoice.Status != 200 {
		return errors.New(res.Invoice.ErrorInfo)
	}

	return nil
}
