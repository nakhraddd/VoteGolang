package blockchain_routes

import (
	"VoteGolang/internals/blockchain"
	"VoteGolang/internals/controller/http/response"
	"VoteGolang/internals/service"
	"net/http"
)

type BlockchainHandler struct {
	Blockchain service.BlockchainService
}

func NewBlockchainHandler(bc service.BlockchainService) *BlockchainHandler {
	return &BlockchainHandler{Blockchain: bc}
}

func (h *BlockchainHandler) GetBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.Blockchain.GetServiceInfo()
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, false, "Failed to get blockchain service info", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, true, "OK", info)
}