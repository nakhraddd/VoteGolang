package service

import (
	"VoteGolang/conf"
	"VoteGolang/internals/domain"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/keys"
)

// TronService implements the BlockchainService interface using the TRON HTTP API
type TronService struct {
	config       *conf.TronConfig
	ownerAddress string
	httpClient   *http.Client
	httpURL      string // e.g., "https://api.shasta.trongrid.io"
}

// NewTronService creates a new connection to a TRON node via HTTP
func NewTronService(config *conf.TronConfig) (BlockchainService, error) {
	if config.NodeURL == "" || config.PrivateKey == "" || config.ContractAddress == "" {
		return nil, fmt.Errorf("TRON config (NodeURL, PrivateKey, ContractAddress) is incomplete")
	}

	// 1. Load wallet address
	privKey, err := keys.GetPrivateKeyFromHex(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid TRON private key: %w", err)
	}
	// Use the .ToECDSA() fix we found
	ownerAddress := address.PubkeyToAddress(*privKey.PubKey().ToECDSA()).String()

	// 2. Set up HTTP client
	// Convert the gRPC URL from .env to the correct HTTP API URL
	var httpURL string
	if strings.Contains(config.NodeURL, "grpc.shasta.trongrid.io") {
		httpURL = "https://api.shasta.trongrid.io"
	} else if strings.Contains(config.NodeURL, "nileex.io") { // For Nile Testnet
		httpURL = "https://api.nileex.io"
	} else {
		httpURL = "https://api.trongrid.io" // Default to mainnet
	}

	log.Println("Successfully connected to TRON HTTP API at", httpURL)
	log.Println("Application wallet address:", ownerAddress)

	return &TronService{
		config:       config,
		ownerAddress: ownerAddress,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		httpURL:      httpURL,
	}, nil
}

// --- Interface Implementation ---

func (s *TronService) LogCandidateCreation(c *domain.Candidate) (*TransactionLog, error) {
	const methodSignature = "logCandidate(uint256,string,string)"
	// Note: We cast c.ID to *big.Int to match uint256
	return s.triggerHTTPContract(methodSignature, new(big.Int).SetUint64(uint64(c.ID)), c.Name, string(c.Type))
}

func (s *TronService) LogCandidateVote(userID uint, candidateID uint, candidateType domain.CandidateType) (*TransactionLog, error) {
	const methodSignature = "logCandidateVote(uint256,uint256,string)"
	return s.triggerHTTPContract(methodSignature, new(big.Int).SetUint64(uint64(userID)), new(big.Int).SetUint64(uint64(candidateID)), string(candidateType))
}

func (s *TronService) LogPetitionCreation(p *domain.Petition) (*TransactionLog, error) {
	const methodSignature = "logPetition(uint256,uint256,string)"
	return s.triggerHTTPContract(methodSignature, new(big.Int).SetUint64(uint64(p.ID)), new(big.Int).SetUint64(uint64(p.UserID)), p.Title)
}

func (s *TronService) LogPetitionVote(userID uint, petitionID uint, voteType domain.VoteType) (*TransactionLog, error) {
	const methodSignature = "logPetitionVote(uint256,uint256,string)"
	return s.triggerHTTPContract(methodSignature, new(big.Int).SetUint64(uint64(userID)), new(big.Int).SetUint64(uint64(petitionID)), string(voteType))
}

func (s *TronService) GetServiceInfo() (map[string]interface{}, error) {
	// Use the HTTP endpoint for node info
	info, err := s.postToTronGrid("/wallet/getnodeinfo", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"service":         "TRON (HTTP)",
		"nodeUrl":         s.config.NodeURL, // <-- FIX HERE
		"contractAddress": s.config.ContractAddress,
		"ownerAddress":    s.ownerAddress,
		"status":          "Connected",
		"currentBlock":    info["block"], // Field from HTTP response
	}, nil
}

// --- TRON Helper Functions ---

// triggerHTTPContract build, signs, and broadcasts a transaction using the HTTP API
func (s *TronService) triggerHTTPContract(functionSignature string, params ...interface{}) (*TransactionLog, error) {
	log.Printf("[TRON HTTP] Calling '%s' with params: %v", functionSignature, params)
	const feeLimit int64 = 10_000_000

	// 1. ABI-encode the parameters (args only, no function selector)
	paramHex, err := abiEncode(functionSignature, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to abi-encode params: %w", err)
	}

	// 2. Build the request payload for TronGrid
	triggerPayload := map[string]interface{}{
		"owner_address":     s.ownerAddress,
		"contract_address":  s.config.ContractAddress,
		"function_selector": functionSignature,
		"parameter":         paramHex,
		"fee_limit":         feeLimit,
		"visible":           true,
	}

	// 3. Post to /triggersmartcontract to get the unsigned transaction
	txExt, err := s.postToTronGrid("/wallet/triggersmartcontract", triggerPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger contract via HTTP: %w", err)
	}

	// 4. Sign the transaction
	// Get private key from hex
	privKey, err := keys.GetPrivateKeyFromHex(s.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Get the raw transaction data to sign
	rawTxDataHex, ok := txExt["transaction"].(map[string]interface{})["raw_data_hex"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid trigger response: missing raw_data_hex")
	}
	rawTxData, err := hex.DecodeString(rawTxDataHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode raw_data_hex: %w", err)
	}

	hash := crypto.Keccak256(rawTxData)
	// Use the .ToECDSA() fix
	signature, err := crypto.Sign(hash, privKey.ToECDSA())
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 5. Broadcast the Signed Transaction
	// Add the signature to the transaction object
	tx := txExt["transaction"].(map[string]interface{})
	tx["signature"] = []string{hex.EncodeToString(signature)}

	broadcastResult, err := s.postToTronGrid("/wallet/broadcasttransaction", tx)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	txID, ok := broadcastResult["txid"].(string)
	if !ok {
		return nil, fmt.Errorf("broadcast did not return txid: %v", broadcastResult)
	}
	log.Printf("[TRON HTTP] Successfully broadcasted TX: %s", txID)

	return &TransactionLog{
		TransactionID: txID,
		Timestamp:     time.Now(),
		ActionType:    functionSignature,
		Details:       params,
	}, nil
}

// postToTronGrid is a generic helper to make HTTP requests
func (s *TronService) postToTronGrid(endpoint string, payload interface{}) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.httpURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add API key if it's in the config
	if s.config.ApiKey != "" { // <-- FIX HERE
		req.Header.Set("TRON-PRO-API-KEY", s.config.ApiKey) // <-- FIX HERE
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TronGrid response: %s", string(bodyBytes))
	}

	// Check for errors in the response
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("TronGrid API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}
	if errStr, ok := result["Error"]; ok {
		return nil, fmt.Errorf("TronGrid API error: %s", errStr)
	}

	return result, nil
}

// abiEncode packs parameters (without the 4-byte selector)
func abiEncode(methodSignature string, args ...interface{}) (string, error) {
	// 1. Extract type names from signature
	// e.g., "logCandidate(uint256,string,string)" -> "uint256,string,string"
	typesStr := methodSignature[strings.Index(methodSignature, "(")+1 : strings.LastIndex(methodSignature, ")")]
	if typesStr == "" {
		return "", nil // No arguments
	}
	typeNames := strings.Split(typesStr, ",")

	// 2. Build abi.Arguments
	var inputs abi.Arguments
	for i, typeName := range typeNames {
		t, err := abi.NewType(typeName, "", nil)
		if err != nil {
			return "", fmt.Errorf("invalid ABI type '%s': %w", typeName, err)
		}
		inputs = append(inputs, abi.Argument{Name: fmt.Sprintf("arg%d", i), Type: t})
	}

	// 3. Pack the arguments
	packedData, err := inputs.Pack(args...)
	if err != nil {
		return "", fmt.Errorf("failed to pack arguments: %w", err)
	}

	// 4. Return as hex string
	return hex.EncodeToString(packedData), nil
}
