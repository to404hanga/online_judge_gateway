package gen

import (
	"log"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGenerateBCrypt(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("generateBCrypt failed: %v", err)
	}
	if len(hash) == 0 {
		t.Errorf("generateBCrypt failed: empty hash")
	}
	log.Println(string(hash))
}
