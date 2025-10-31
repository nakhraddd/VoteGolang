package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// Transaction represents a record of an action, e.g., a vote or a new candidate.
type Transaction struct {
	Type        string      `json:"type"`
	Payload     interface{} `json:"payload"`
	Timestamp   time.Time   `json:"timestamp"`
	Description string      `json:"description"`
}

// Block represents a single block in the blockchain.
type Block struct {
	Index        int           `json:"index"`
	Timestamp    time.Time     `json:"timestamp"`
	Transaction  Transaction   `json:"transaction"`
	PrevHash     string        `json:"prev_hash"`
	Hash         string        `json:"hash"`
	Nonce        int           `json:"nonce"`
	Difficulty   int           `json:"difficulty"`
}

// Blockchain is a series of validated Blocks.
type Blockchain struct {
	Chain      []*Block `json:"chain"`
	Difficulty int      `json:"difficulty"`
}

// NewBlockchain creates a new Blockchain with a genesis Block.
func NewBlockchain(difficulty int) *Blockchain {
	genesisBlock := &Block{
		Index:        0,
		Timestamp:    time.Now(),
		Transaction:  Transaction{Type: "genesis", Payload: "Genesis Block"},
		PrevHash:     "",
		Difficulty:   difficulty,
	}
	genesisBlock.Hash = genesisBlock.calculateHash()

	return &Blockchain{
		Chain:      []*Block{genesisBlock},
		Difficulty: difficulty,
	}
}

// calculateHash generates the hash for a Block using a standardized format.
func (b *Block) calculateHash() string {
	// Use a consistent timestamp format
	timestamp := b.Timestamp.Format(time.RFC3339Nano)

	// Marshal the transaction payload for a consistent representation
	transactionBytes, err := json.Marshal(b.Transaction)
	if err != nil {
		// In a real application, this should be handled more gracefully
		return ""
	}

	// Include all relevant fields, including the nonce
	record := fmt.Sprintf("%d%s%s%s%d%d", b.Index, timestamp, string(transactionBytes), b.PrevHash, b.Nonce, b.Difficulty)
	h := sha256.New()
	h.Write([]byte(record))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MineBlock finds a hash that satisfies the difficulty requirement (proof-of-work).
func (b *Block) MineBlock(difficulty int) {
	target := make([]byte, difficulty) // Creates a slice of `difficulty` zeros
	for {
		b.Hash = b.calculateHash()
		// Compare the first `difficulty` bytes of the hash with the target
		if b.Hash[:difficulty] == string(target) {
			break
		}
		b.Nonce++
	}
	fmt.Printf("Block mined: %s\n", b.Hash)
}


// AddBlock creates a new block, mines it, and then adds it to the blockchain.
func (bc *Blockchain) AddBlock(transaction Transaction) {
	prevBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := &Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now(),
		Transaction:  transaction,
		PrevHash:     prevBlock.Hash,
		Difficulty:   bc.Difficulty,
		Nonce:        0, // Start nonce from 0 for mining
	}
	newBlock.MineBlock(bc.Difficulty) // Perform proof-of-work
	bc.Chain = append(bc.Chain, newBlock)
}

// GetChain returns the full blockchain.
func (bc *Blockchain) GetChain() []*Block {
	return bc.Chain
}