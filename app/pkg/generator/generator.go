package generator

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/tyler-smith/go-bip39"
)

func SeedPhraseGenerator() string {
	// Step 1: Generate 256 bits of entropy
	entropy, err := bip39.NewEntropy(256) // 256 bits
	if err != nil {
		panic(err)
	}

	// Step 2: Generate mnemonic from entropy
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic(err)
	}

	// Output the mnemonic
	fmt.Println("Mnemonic (24 words):", mnemonic)
	return mnemonic
}

func CreateBTCHash() string {
	key := make([]byte, 8)
	_, err := rand.Read(key)
	if err != nil {
		// handle error here
		fmt.Println(err)
	}
	str := hex.EncodeToString(key)
	return str
}
