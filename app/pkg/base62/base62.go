package base62

import (
	"math"
	"strings"
)

const characters = "Aa1Bb2Cc3Dd4Ee5Ff6Gg7Hh8Ii9Jj0KkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"

func Encode(num int32) string {

	result := make([]byte, 0)

	for num > 0 {
		remainder := num % 62
		result = append([]byte{characters[remainder]}, result...)
		num = num / 62
	}

	// Rellenar con ceros a la izquierda para tener 6 caracteres
	padding := 6 - len(result)
	for i := 0; i < padding; i++ {
		result = append([]byte{characters[0]}, result...)
	}

	return string(result)
}

func Decode(str string) int {
	result := 0
	strLen := len(str)

	for i, char := range str {
		power := strLen - i - 1
		result += strings.IndexByte(characters, byte(char)) * int(math.Pow(62, float64(power)))
	}

	return result
}
