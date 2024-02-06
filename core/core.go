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

var ProgramId = solana.MustPublicKeyFromBase58("DeCW2vrjkkproab8aRGyzx4YACC9Rt2NNJD3TA4mz7WR")
var SwapAuthority = solana.MustPublicKeyFromBase58("3UmYo43UWktvk247FHCpm1nAfc3s5e4Qk2vM9qjyTadC")
var SwapPool = solana.MustPublicKeyFromBase58("CNii9WF5t5eHwcFX77RiCnqqNs6V2nsAedGZQEfy73H7")
var SwapNgnAta = solana.MustPublicKeyFromBase58("8viMHJ2JMmiM6WAxY4bKpT6nhqRPBmkARn8dQHyA1jjp")
var SwapUsdAta = solana.MustPublicKeyFromBase58("CLiARrhtW3dcudjhWNLZPivAVzMDhMjaWjti6Bw1g2MV")

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
