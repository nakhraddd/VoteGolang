package blockchain_routes

import (
	"VoteGolang/internals/blockchain"
	"VoteGolang/internals/controller/http/response"
	"net/http"
)

type BlockchainHandler struct {
	Blockchain *blockchain.Blockchain
}

func NewBlockchainHandler(bc *blockchain.Blockchain) *BlockchainHandler {
	return &BlockchainHandler{Blockchain: bc}
}

func (h *BlockchainHandler) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, true, "OK", h.Blockchain.GetChain())
}