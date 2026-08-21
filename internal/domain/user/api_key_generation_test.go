package user

import (
	"strings"
	"testing"
)

func TestAPIKeyGenerator_Generate(t *testing.T) {
	g := NewAPIKeyGenerator()

	fullKey, keyID, err := g.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(fullKey, KeyPrefix+"_"+KeyVersion+"_") {
		t.Errorf("expected key to have prefix %s_%s_, got: %s", KeyPrefix, KeyVersion, fullKey)
	}

	if len(keyID) != 16 {
		t.Errorf("expected keyID length 16, got: %d", len(keyID))
	}

	// Verify keyID matches first 16 chars of random part
	prefix, version, randomPart, err := g.ParseKey(fullKey)
	if err != nil {
		t.Fatalf("failed to parse key: %v", err)
	}

	if prefix != KeyPrefix {
		t.Errorf("expected prefix %s, got %s", KeyPrefix, prefix)
	}

	if version != KeyVersion {
		t.Errorf("expected version %s, got %s", KeyVersion, version)
	}

	if randomPart[:16] != keyID {
		t.Errorf("expected keyID %s to match prefix of randomPart %s", keyID, randomPart)
	}
}

func TestAPIKeyGenerator_ValidateFormat(t *testing.T) {
	g := NewAPIKeyGenerator()

	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{
			name:  "invalid length",
			key:   "tsr_v1_short",
			valid: false,
		},
		{
			name:  "invalid prefix",
			key:   "bad_v1_abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG",
			valid: false,
		},
		{
			name:  "invalid parts",
			key:   "tsr_v1",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if g.ValidateFormat(tt.key) != tt.valid {
				t.Errorf("expected validity %v for key %s, got %v", tt.valid, tt.key, !tt.valid)
			}
		})
	}

	// Test with a real generated key
	realKey, _, err := g.Generate()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	if !g.ValidateFormat(realKey) {
		t.Errorf("expected real generated key %s to be valid format", realKey)
	}
}
