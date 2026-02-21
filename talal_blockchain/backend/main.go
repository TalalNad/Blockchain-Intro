package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Block struct {
	Index      int      `json:"index"`
	Timestamp  string   `json:"timestamp"`
	Txs        []string `json:"txs"`        // transactions as strings
	MerkleRoot string   `json:"merkleRoot"` // will be proper merkle root later
	PrevHash   string   `json:"prevHash"`
	Hash       string   `json:"hash"`
	Nonce      int      `json:"nonce"`
}

type Blockchain struct {
	Name        string   `json:"name"`
	Chain       []Block  `json:"chain"`
	PendingTxs  []string `json:"pendingTxs"`
	Difficulty  int      `json:"difficulty"`
}

var (
	bc *Blockchain
	mu sync.Mutex
)

type addTxRequest struct {
	Tx string `json:"tx"`
}

type messageResponse struct {
	Message string `json:"message"`
}

type searchResult struct {
	BlockIndex int    `json:"blockIndex"`
	Tx         string `json:"tx"`
}

// calculateHash hashes block header fields (not the whole block JSON)
func calculateHash(index int, timestamp string, merkleRoot string, prevHash string, nonce int) string {
	record := fmt.Sprintf("%d|%s|%s|%s|%d", index, timestamp, merkleRoot, prevHash, nonce)
	sum := sha256.Sum256([]byte(record))
	return hex.EncodeToString(sum[:])
}

// merkleRoot computes a Merkle root for the given transactions.
//
// Rules:
// - Each tx string is hashed with SHA-256 to form the leaf level.
// - Parent nodes are SHA-256(left || right).
// - If a level has an odd number of nodes, the last node is duplicated.
// - If there are no transactions, the Merkle root is SHA-256(""), i.e. the hash of empty bytes.
func merkleRoot(txs []string) string {
	// Empty tx list: define root as hash of empty bytes
	if len(txs) == 0 {
		sum := sha256.Sum256([]byte{})
		return hex.EncodeToString(sum[:])
	}

	// Build leaf level
	level := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		h := sha256.Sum256([]byte(tx))
		b := make([]byte, len(h[:]))
		copy(b, h[:])
		level = append(level, b)
	}

	// Build tree up to the root
	for len(level) > 1 {
		// If odd, duplicate last
		if len(level)%2 == 1 {
			dup := make([]byte, len(level[len(level)-1]))
			copy(dup, level[len(level)-1])
			level = append(level, dup)
		}

		next := make([][]byte, 0, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			combined := append(level[i], level[i+1]...)
			sum := sha256.Sum256(combined)
			b := make([]byte, len(sum[:]))
			copy(b, sum[:])
			next = append(next, b)
		}
		level = next
	}

	return hex.EncodeToString(level[0])
}

func newGenesisBlock(roll string) Block {
	txs := []string{roll} // REQUIRED: first tx in genesis = roll number
	ts := time.Now().UTC().Format(time.RFC3339)
	mr := merkleRoot(txs)

	gen := Block{
		Index:      0,
		Timestamp:  ts,
		Txs:        txs,
		MerkleRoot: mr,
		PrevHash:   "0",
		Nonce:      0,
	}
	gen.Hash = calculateHash(gen.Index, gen.Timestamp, gen.MerkleRoot, gen.PrevHash, gen.Nonce)
	return gen
}

func NewBlockchain(name string, roll string, difficulty int) *Blockchain {
	if difficulty <= 0 {
		difficulty = 3 // sensible default; PoW step will use this
	}
	bc := &Blockchain{
		Name:       name,
		Chain:      []Block{newGenesisBlock(roll)},
		PendingTxs: make([]string, 0),
		Difficulty: difficulty,
	}
	return bc
}

// AddTransaction adds a non-empty transaction string to the pending pool.
func (bc *Blockchain) AddTransaction(tx string) error {
	tx = strings.TrimSpace(tx)
	if tx == "" {
		return errors.New("transaction cannot be empty")
	}
	bc.PendingTxs = append(bc.PendingTxs, tx)
	return nil
}

// BuildNextBlock constructs the next block from pending transactions (NOT mined yet).
// It clears the pending pool after building the block.
func (bc *Blockchain) BuildNextBlock() (Block, error) {
	if len(bc.PendingTxs) == 0 {
		return Block{}, errors.New("no pending transactions to put in a block")
	}

	prev := bc.Chain[len(bc.Chain)-1]
	index := prev.Index + 1
	ts := time.Now().UTC().Format(time.RFC3339)

	// copy pending txs into block
	txs := make([]string, len(bc.PendingTxs))
	copy(txs, bc.PendingTxs)

	mr := merkleRoot(txs)
	b := Block{
		Index:      index,
		Timestamp:  ts,
		Txs:        txs,
		MerkleRoot: mr,
		PrevHash:   prev.Hash,
		Nonce:      0,
	}
	b.Hash = calculateHash(b.Index, b.Timestamp, b.MerkleRoot, b.PrevHash, b.Nonce)

	// clear pending after building
	bc.PendingTxs = bc.PendingTxs[:0]

	return b, nil
}

// AppendBlock appends a block to the chain (no validation yet; we will add validation later).
func (bc *Blockchain) AppendBlock(b Block) {
	bc.Chain = append(bc.Chain, b)
}

func hasLeadingZeros(hash string, difficulty int) bool {
	if difficulty <= 0 {
		return true
	}
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// MineBlock performs Proof-of-Work on a block by incrementing the nonce until
// the hash has `difficulty` leading zeros.
func MineBlock(b Block, difficulty int) Block {
	for {
		b.Hash = calculateHash(b.Index, b.Timestamp, b.MerkleRoot, b.PrevHash, b.Nonce)
		if hasLeadingZeros(b.Hash, difficulty) {
			return b
		}
		b.Nonce++
	}
}

// MineNextBlock builds the next block from pending txs, mines it, and appends it.
func (bc *Blockchain) MineNextBlock() (Block, error) {
	b, err := bc.BuildNextBlock()
	if err != nil {
		return Block{}, err
	}

	start := time.Now()
	mined := MineBlock(b, bc.Difficulty)
	elapsed := time.Since(start)

	bc.AppendBlock(mined)

	fmt.Printf("\n⛏️  Mined block %d with difficulty %d in %s\n", mined.Index, bc.Difficulty, elapsed)
	fmt.Printf("  Nonce: %d\n", mined.Nonce)
	fmt.Printf("  Hash:  %s\n", mined.Hash)

	return mined, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, messageResponse{Message: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, messageResponse{Message: "ok"})
}

func handleChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, messageResponse{Message: "method not allowed"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	writeJSON(w, http.StatusOK, bc)
}

func handlePending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, messageResponse{Message: "method not allowed"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"pendingTxs": bc.PendingTxs})
}

func handleAddTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, messageResponse{Message: "method not allowed"})
		return
	}

	var req addTxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, messageResponse{Message: "invalid JSON body"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if err := bc.AddTransaction(req.Tx); err != nil {
		writeJSON(w, http.StatusBadRequest, messageResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message":    "transaction added",
		"pendingTxs": bc.PendingTxs,
	})
}

func handleMine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, messageResponse{Message: "method not allowed"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	mined, err := bc.MineNextBlock()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, messageResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, mined)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, messageResponse{Message: "method not allowed"})
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusBadRequest, messageResponse{Message: "missing query param q"})
		return
	}
	qLower := strings.ToLower(q)

	mu.Lock()
	defer mu.Unlock()

	results := make([]searchResult, 0)
	for _, b := range bc.Chain {
		for _, tx := range b.Txs {
			if strings.Contains(strings.ToLower(tx), qLower) {
				results = append(results, searchResult{BlockIndex: b.Index, Tx: tx})
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"query":   q,
		"count":   len(results),
		"results": results,
	})
}

func main() {
	name := "Talal Nadeem"
	roll := "22L-6679"

	// Difficulty is number of leading zeros required in the hash.
	// You can increase this later; 3 is a reasonable starting value for demos.
	bc = NewBlockchain(name, roll, 3)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", withCORS(handleHealth))
	mux.HandleFunc("/chain", withCORS(handleChain))
	mux.HandleFunc("/pending", withCORS(handlePending))
	mux.HandleFunc("/tx", withCORS(handleAddTx))
	mux.HandleFunc("/mine", withCORS(handleMine))
	mux.HandleFunc("/search", withCORS(handleSearch))

	addr := ":8080"
	fmt.Println("✅ Server running at http://localhost" + addr)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /chain")
	fmt.Println("  GET  /pending")
	fmt.Println("  POST /tx      {\"tx\":\"Alice -> Bob : 5\"}")
	fmt.Println("  POST /mine")
	fmt.Println("  GET  /search?q=Bob")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Println("server error:", err)
	}
}