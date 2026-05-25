package myutils

import "testing"

func TestHashPasswordAndCompare(t *testing.T) {
	hash := HashPassword("secret")
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "secret" {
		t.Fatal("hash should not equal raw password")
	}
	if !CompareHashAndPassword("secret", hash) {
		t.Fatal("expected password to match hash")
	}
	if CompareHashAndPassword("wrong", hash) {
		t.Fatal("expected wrong password not to match hash")
	}
}
