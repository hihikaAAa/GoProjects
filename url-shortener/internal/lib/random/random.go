package random

import(
	"strings"
	"math/rand"
)

func NewUniqueRandomString(aliasLength int) string{
	var sb strings.Builder
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	for range aliasLength{
		sb.WriteRune(letters[rand.Intn(len(letters))])
	}
	return sb.String()
}