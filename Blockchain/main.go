package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	wallet "github.com/sunnyradadiya/Github/GO-WALLET/Wallet"
)

type Blockchain struct {
	blocks []*Block
}

type Block struct {
	Hash             []byte
	Token            int
	PrevHash         []byte
	SenderAddress    string
	RecipientAddress string
	Timestamp        string
}

var wallets wallet.Wallets
var chain Blockchain

// Generate hash for new block
func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.PrevHash, []byte(b.Timestamp)}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

// check newly created block is valid
// new block hash and previous block prevhash should match
func isBlockValid(newBlock, oldBlock *Block) bool {
	res := bytes.Compare(oldBlock.Hash, newBlock.PrevHash)
	if res != 0 {
		return false
	}
	return true
}

// create new block with all parameters
func CreateBlock(token int, prevHash []byte, senderAddress string, recipientAddress string) *Block {
	t := time.Now()
	block := &Block{[]byte{}, token, prevHash, senderAddress, recipientAddress, t.String()}
	block.DeriveHash()
	return block
}

// Add block to blockchain
func (chain *Blockchain) AddBlock(token int, senderAddress string, recipientAddress string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	new := CreateBlock(token, prevBlock.Hash, senderAddress, recipientAddress)
	if isBlockValid(new, prevBlock) {
		chain.blocks = append(chain.blocks, new)
	}
}

// Genesis block
func Genesis() *Block {
	return CreateBlock(0, []byte{}, "", "")
}

// Initialize blockchain
func InitBlockchain() *Blockchain {
	return &Blockchain{[]*Block{Genesis()}}
}

////////// APIs //////////

func addWallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	address := wallets.AddWallet()
	_ = json.NewDecoder(r.Body).Decode(address)
	addresses := wallet.GetAllAddresses()
	addresses = append(addresses, address)
	json.NewEncoder(w).Encode(&address)
}

func getAllWalletAddresses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	addresses := wallet.GetAllAddresses()
	json.NewEncoder(w).Encode(addresses)
}

func getAllWalletDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallets)
}

func TransferToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	token, _ := strconv.Atoi(vars["val"])
	sender := vars["sender"]
	recipient := vars["recipient"]
	_ = json.NewDecoder(r.Body).Decode(chain)
	if wallets.Wallets[sender].Token >= token {
		wallets.Wallets[sender].Token = wallets.Wallets[sender].Token - token
		wallets.Wallets[recipient].Token = wallets.Wallets[recipient].Token + token
		chain.AddBlock(token, sender, recipient)
	}
	json.NewEncoder(w).Encode(&chain.blocks)
}

func main() {
	chain.blocks = append(chain.blocks, Genesis())
	fmt.Println(chain)

	router := mux.NewRouter()
	router.HandleFunc("/addwallet", addWallet).Methods("POST")
	router.HandleFunc("/getAllWalletAddresses", getAllWalletAddresses).Methods("GET")
	router.HandleFunc("/getAllWalletDetails", getAllWalletDetails).Methods("GET")
	router.HandleFunc("/transferToken/{val}/{sender}/{recipient}", TransferToken).Methods("POST")
	http.ListenAndServe(":8000", router)
}
