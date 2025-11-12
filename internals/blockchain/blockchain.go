package blockchain

import (
	"VoteGolang/internals/app/logging"
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
	Index       int         `json:"index"`
	Timestamp   time.Time   `json:"timestamp"`
	Transaction Transaction `json:"transaction"`
	PrevHash    string      `json:"prev_hash"`
	Hash        string      `json:"hash"`
	Nonce       int         `json:"nonce"`
	Difficulty  int         `json:"difficulty"`
}

// Blockchain is a series of validated Blocks.
type Blockchain struct {
	Chain       []*Block             `json:"chain"`
	Difficulty  int                  `json:"difficulty"`
	KafkaLogger *logging.KafkaLogger `json:"-"`
}

// NewBlockchain creates a new Blockchain with a genesis Block.
func NewBlockchain(difficulty int, kafkaLogger *logging.KafkaLogger) *Blockchain {
	genesisBlock := &Block{
		Index:       0,
		Timestamp:   time.Now(),
		Transaction: Transaction{Type: "genesis", Payload: "Genesis Block"},
		PrevHash:    "",
		Difficulty:  difficulty,
	}
	genesisBlock.Hash = genesisBlock.calculateHash()

	kafkaLogger.Log("INFO", fmt.Sprintf("Genesis block created with hash %s", genesisBlock.Hash))

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

//func (b *Block) MineBlock(difficulty int) {
//	target := make([]byte, difficulty) // Creates a slice of `difficulty` zeros
//	for {
//		b.Hash = b.calculateHash()
//		// Compare the first `difficulty` bytes of the hash with the target
//		if b.Hash[:difficulty] == string(target) {
//			break
//		}
//		b.Nonce++
//	}
//	fmt.Printf("Block mined: %s\n", b.Hash)
//}

func (b *Block) MineBlock(difficulty int, kafkaLogger *logging.KafkaLogger) {
	target := ""
	for i := 0; i < difficulty; i++ {
		target += "0"
	}
	for {
		b.Hash = b.calculateHash()
		if b.Hash[:difficulty] == target {
			break
		}
		b.Nonce++
	}
	kafkaLogger.Log("INFO", fmt.Sprintf("⛏️ Starting mining for block %d with difficulty %d", b.Index, difficulty))
}

// AddBlock creates a new block, mines it, and then adds it to the blockchain.
func (bc *Blockchain) AddBlock(transaction Transaction) {
	prevBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := &Block{
		Index:       prevBlock.Index + 1,
		Timestamp:   time.Now(),
		Transaction: transaction,
		PrevHash:    prevBlock.Hash,
		Difficulty:  bc.Difficulty,
		Nonce:       0, // Start nonce from 0 for mining
	}

	bc.KafkaLogger.Log("INFO", fmt.Sprintf("Creating new block %d with transaction type '%s'", newBlock.Index, transaction.Type))

	newBlock.MineBlock(bc.Difficulty, bc.KafkaLogger) // Perform proof-of-work
	bc.Chain = append(bc.Chain, newBlock)

	bc.KafkaLogger.Log("INFO", fmt.Sprintf("Block %d added to chain with hash %s", newBlock.Index, newBlock.Hash))
}

// GetChain returns the full blockchain.
func (bc *Blockchain) GetChain() []*Block {
	bc.KafkaLogger.Log("DEBUG", fmt.Sprintf("Blockchain retrieved with %d blocks", len(bc.Chain)))
	return bc.Chain
}
