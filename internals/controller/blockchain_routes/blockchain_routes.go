package blockchain_routes

import (
	"net/http"
)

func RegisterBlockchainRoutes(mux *http.ServeMux, handler *BlockchainHandler) {
	mux.HandleFunc("/blockchain", handler.GetBlockchainInfo)
}
