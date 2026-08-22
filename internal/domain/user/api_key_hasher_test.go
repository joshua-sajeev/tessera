package user

import (
	"testing"
)

func TestKeyHasher_HashAndVerify(t *testing.T) {
	h := NewKeyHasher()

	key := "tsr_v1_someSecureRandomKeyStringGoesHere"

	hash, err := h.Hash(key)
	if err != nil {
		t.Fatalf("failed to hash key: %v", err)
	}

	valid, err := h.Verify(key, hash)
	if err != nil {
		t.Fatalf("unexpected error verifying key: %v", err)
	}
	if !valid {
		t.Errorf("expected key to be verified as valid")
	}

	// Verify that a wrong key does not match
	valid, err = h.Verify("wrongKey", hash)
	if err != nil {
		t.Fatalf("unexpected error verifying wrong key: %v", err)
	}
	if valid {
		t.Errorf("expected wrong key to be invalid")
	}
}

func TestKeyHasher_VerifyInvalidFormat(t *testing.T) {
	h := NewKeyHasher()

	invalidHashes := []string{
		"",
		"argon2id$v=19$m=65536,t=1,p=4$salt$hash", // needs starting $
		"$argon2i$v=19$m=65536,t=1,p=4$salt$hash",  // wrong algorithm
		"$argon2id$v=99$m=65536,t=1,p=4$salt$hash", // wrong version
		"$argon2id$v=19$m=0,t=1,p=4$salt$hash",     // invalid parameters
		"$argon2id$v=19$m=65536,t=1,p=4$salt",      // too few parts
	}

	for _, hash := range invalidHashes {
		valid, err := h.Verify("key", hash)
		if err == nil {
			t.Errorf("expected error for hash: %s", hash)
		}
		if valid {
			t.Errorf("expected invalid result for hash: %s", hash)
		}
	}
}
