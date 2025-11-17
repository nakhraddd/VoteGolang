package service

import (
	"VoteGolang/conf"
	"VoteGolang/internals/domain"
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BnbService implements the BlockchainService interface using an EVM-compatible JSON-RPC client
type BnbService struct {
	config          *conf.BnbConfig // Assumes a new config struct for BNB
	client          *ethclient.Client
	ownerAddress    common.Address
	privateKey      *ecdsa.PrivateKey
	contractAddress common.Address
	contractABI     abi.ABI
	chainID         *big.Int
}

// NewBnbService creates a new connection to a BNB node via JSON-RPC
func NewBnbService(config *conf.BnbConfig) (BlockchainService, error) {
	if config.NodeURL == "" || config.PrivateKey == "" || config.ContractAddress == "" || config.ChainID == 0 {
		return nil, fmt.Errorf("BNB config (NodeURL, PrivateKey, ContractAddress, ChainID) is incomplete")
	}

	// 1. Connect to the EVM node
	client, err := ethclient.Dial(config.NodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BNB node: %w", err)
	}

	// 2. Load private key and derive owner address
	privateKeyHex := strings.TrimPrefix(strings.ToLower(config.PrivateKey), "0x")
	privKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid BNB private key: %w", err)
	}

	pubKey := privKey.Public()
	pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	ownerAddress := crypto.PubkeyToAddress(*pubKeyECDSA)

	// 3. Parse Contract Address
	contractAddress := common.HexToAddress(config.ContractAddress)

	// 4. Parse Contract ABI (same as before)
	const contractABIJSON = `[{"name":"logCandidate","type":"function","inputs":[{"name":"candidateId","type":"uint256"},{"name":"name","type":"string"},{"name":"candidateType","type":"string"}]}, {"name":"logCandidateVote","type":"function","inputs":[{"name":"userId","type":"uint256"},{"name":"candidateId","type":"uint256"},{"name":"candidateType","type":"string"}]}, {"name":"logPetition","type":"function","inputs":[{"name":"petitionId","type":"uint256"},{"name":"userId","type":"uint256"},{"name":"title","type":"string"}]}, {"name":"logPetitionVote","type":"function","inputs":[{"name":"userId","type":"uint256"},{"name":"petitionId","type":"uint256"},{"name":"voteType","type":"string"}]}]`

	parsedABI, err := abi.JSON(strings.NewReader(contractABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	chainID := big.NewInt(config.ChainID)
	log.Println("Successfully connected to BNB RPC at", config.NodeURL)
	log.Println("Application wallet address:", ownerAddress.Hex())
	log.Println("Target contract address:", contractAddress.Hex())
	log.Println("Using Chain ID:", chainID)

	return &BnbService{
		config:          config,
		client:          client,
		ownerAddress:    ownerAddress,
		privateKey:      privKey,
		contractAddress: contractAddress,
		contractABI:     parsedABI,
		chainID:         chainID,
	}, nil
}

// --- Interface Implementation ---

func (s *BnbService) LogCandidateCreation(c *domain.Candidate) (*TransactionLog, error) {
	const methodSignature = "logCandidate(uint256,string,string)"
	return s.sendEVMTx(methodSignature, new(big.Int).SetUint64(uint64(c.ID)), c.Name, string(c.Type))
}

func (s *BnbService) LogCandidateVote(userID uint, candidateID uint, candidateType domain.CandidateType) (*TransactionLog, error) {
	const methodSignature = "logCandidateVote(uint256,uint256,string)"
	return s.sendEVMTx(methodSignature, new(big.Int).SetUint64(uint64(userID)), new(big.Int).SetUint64(uint64(candidateID)), string(candidateType))
}

func (s *BnbService) LogPetitionCreation(p *domain.Petition) (*TransactionLog, error) {
	const methodSignature = "logPetition(uint256,uint256,string)"
	return s.sendEVMTx(methodSignature, new(big.Int).SetUint64(uint64(p.ID)), new(big.Int).SetUint64(uint64(p.UserID)), p.Title)
}

func (s *BnbService) LogPetitionVote(userID uint, petitionID uint, voteType domain.VoteType) (*TransactionLog, error) {
	const methodSignature = "logPetitionVote(uint256,uint256,string)"
	return s.sendEVMTx(methodSignature, new(big.Int).SetUint64(uint64(userID)), new(big.Int).SetUint64(uint64(petitionID)), string(voteType))
}

func (s *BnbService) GetServiceInfo() (map[string]interface{}, error) {
	header, err := s.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"service":         "BNB Smart Chain (RPC)",
		"nodeUrl":         s.config.NodeURL,
		"contractAddress": s.contractAddress.Hex(),
		"ownerAddress":    s.ownerAddress.Hex(),
		"status":          "Connected",
		"currentBlock":    header.Number.String(),
	}, nil
}

// --- BNB/EVM Helper Function ---

// sendEVMTx builds, signs, and broadcasts a transaction to an EVM chain
func (s *BnbService) sendEVMTx(functionSignature string, params ...interface{}) (*TransactionLog, error) {
	log.Printf("[BNB RPC] Calling '%s' with params: %v", functionSignature, params)
	methodName := functionSignature[:strings.Index(functionSignature, "(")]

	// 1. Pack transaction data
	packedData, err := s.contractABI.Pack(methodName, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack arguments for method %s: %w", methodName, err)
	}

	// 2. Get nonce
	nonce, err := s.client.PendingNonceAt(context.Background(), s.ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending nonce: %w", err)
	}

	// 3. Get gas price
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %w", err)
	}

	// 4. Estimate gas limit
	msg := ethereum.CallMsg{
		From: s.ownerAddress,
		To:   &s.contractAddress,
		Data: packedData,
	}
	gasLimit, err := s.client.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w. Check contract and params", err)
	}

	// 5. Create the transaction object (value is 0)
	tx := types.NewTransaction(nonce, s.contractAddress, big.NewInt(0), gasLimit, gasPrice, packedData)

	// 6. Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.chainID), s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 7. Send the transaction
	err = s.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	log.Printf("[BNB RPC] Successfully broadcasted TX: %s | From: %s | Contract: %s | Function: %s",
		signedTx.Hash().Hex(),
		s.ownerAddress.Hex(),
		s.contractAddress.Hex(),
		functionSignature)

	// 8. Wait for the transaction to be mined
	log.Printf("[BNB RPC] Waiting for TX %s to be mined...", signedTx.Hash().Hex())
	receipt, err := bind.WaitMined(context.Background(), s.client, signedTx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for tx %s to be mined: %w", signedTx.Hash().Hex(), err)
	}

	if receipt.Status == 0 {
		// Transaction reverted
		return nil, fmt.Errorf("transaction %s reverted by EVM", signedTx.Hash().Hex())
	}

	// 9. Calculate fee (Fee = GasUsed * EffectiveGasPrice)
	// Note: We use Wei, not SUN.
	feePaid := new(big.Int).Mul(receipt.EffectiveGasPrice, big.NewInt(int64(receipt.GasUsed)))

	log.Printf("[BNB RPC] TX %s confirmed. Block: %d | Fee Paid: %s Wei (%.18f BNB)",
		receipt.TxHash.Hex(),
		receipt.BlockNumber,
		feePaid.String(),
		new(big.Float).Quo(new(big.Float).SetInt(feePaid), big.NewFloat(1e18)))

	return &TransactionLog{
		TransactionID: receipt.TxHash.Hex(),
		Timestamp:     time.Now(),
		ActionType:    functionSignature,
		Details:       params,
		FeeWei:        feePaid.Int64(), // Use the updated struct field
	}, nil
}
