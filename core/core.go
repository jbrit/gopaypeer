package core

import (
	"crypto/rand"
	"io"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var Client = rpc.New("https://rpc.ironforge.network/devnet?apiKey=01HNBAZBY7BKD27PNJC6FV26YR")
var NgnMint = solana.MustPublicKeyFromBase58("NGNTfR7uP1z678g1PMdad4ds4r5jYFbKQe1KortAKg4")
var UsdMint = solana.MustPublicKeyFromBase58("USDu4C6jm4Cxeq275QfHTEfXFjHpNxycmWYqNqMfTxb")

func GetRandomNumberString(length int) (string, error) {
	numbers := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	cardNumberSuffix := make([]byte, length)
	_, err := io.ReadAtLeast(rand.Reader, cardNumberSuffix, length)
	if err != nil {
		return "", err
	}
	for i := 0; i < len(cardNumberSuffix); i++ {
		cardNumberSuffix[i] = numbers[int(cardNumberSuffix[i])%10]
	}
	return string(cardNumberSuffix), nil
}
