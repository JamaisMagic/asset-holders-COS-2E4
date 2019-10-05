package handlers

import (
	"net/http"
	"encoding/json"
	"fmt"
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

func LikePost(write http.ResponseWriter, request *http.Request) {
	var voteReq VoteReq
	err := json.NewDecoder(request.Body).Decode(&voteReq)

	if err != nil {
		http.Error(write, err.Error(), 400)
		return
	}

	fmt.Println(voteReq.AccountName)

	var voteRes VoteRes

	voteRes.Status = 1
	voteRes.Message = "Success"

	output, err := json.Marshal(voteRes)

	if err != nil {
		http.Error(write, err.Error(), 500)
	}

	write.Header().Set("Content-Type", "application/json")
	write.Write(output)
}
