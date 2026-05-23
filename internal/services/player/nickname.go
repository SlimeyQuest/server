package player

import (
	"crypto/rand"
	"math/big"
)

const nicknameSuffixAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// DefaultNickname returns a lightweight game-style default name, e.g. Slime-A3F9.
func DefaultNickname() string {
	return "Slime-" + randomSuffix(4)
}

func randomSuffix(length int) string {
	out := make([]byte, length)
	max := big.NewInt(int64(len(nicknameSuffixAlphabet)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			out[i] = nicknameSuffixAlphabet[i%len(nicknameSuffixAlphabet)]
			continue
		}
		out[i] = nicknameSuffixAlphabet[n.Int64()]
	}
	return string(out)
}

// ValidateDefaultNicknamePattern is used by tests to assert nickname format.
func ValidateDefaultNicknamePattern(nickname string) bool {
	if len(nickname) != len("Slime-")+4 {
		return false
	}
	if nickname[:6] != "Slime-" {
		return false
	}
	for _, ch := range nickname[6:] {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}
