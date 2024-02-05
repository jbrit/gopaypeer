package models

import (
	"crypto/rand"
	"io"
)

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
