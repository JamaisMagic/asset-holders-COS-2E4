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

// conn, _ := rpc.Dial("https://mainnode.contentos.io")
var conn, _ = rpc.Dial("34.203.85.235:8080")
var rpcClient = grpcpb.NewApiServiceClient(conn)
var mChainIdName string = "dev"


func VotePost(write http.ResponseWriter, request *http.Request) {
	var voteReq VoteReq
	err := json.NewDecoder(request.Body).Decode(&voteReq)

	if err != nil {
		http.Error(write, err.Error(), 400)
		return
	}

	fmt.Println(voteReq.AccountName)

	postIdInt64, _ := strconv.ParseUint(voteReq.PostId, 10, 64)

	fmt.Println("go vote: ", voteReq.AccountName, voteReq.PrivateKey, postIdInt64)

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

	fmt.Println("signedTx", signedTx)

	errReq := txRequest(signedTx)

	if errReq != nil {
		fmt.Println("txRequestett", errReq.Error())
		return errReq
	}

	return nil
}

func signTx(privateKey string, ops ...interface{}) (*prototype.SignedTransaction, error) {
	privKey := &prototype.PrivateKeyType{}
	pk, err := prototype.PrivateKeyFromWIF(privateKey)
	if err != nil {
		return nil, err
	}
	privKey = pk

	chainState, err := getChainState()

	if err != nil {
		return nil, err
	}

	refBlockPrefix := binary.BigEndian.Uint32(chainState.Dgpo.HeadBlockId.Hash[8:12])
	// occupant implement
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
		return nil, err
	}

	return &signTx, nil
}

func getChainState() (*grpcpb.ChainState, error) {
	req := &grpcpb.NonParamsRequest{}
	resp, err := rpcClient.GetChainState(context.Background(), req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("res: nil")
	}
	return resp.State, nil
}

func txRequest(signedTx *prototype.SignedTransaction) error {
	req := &grpcpb.BroadcastTrxRequest{Transaction: signedTx, Finality: false}

	res, err := rpcClient.BroadcastTrx(context.Background(), req)
	if err != nil {
		return err
	}

	if res == nil {
		return errors.New("BroadcastTrx res nil")
	}

	if res.Invoice.Status != 200 {
		return errors.New(res.Invoice.ErrorInfo)
	}

	return nil
}