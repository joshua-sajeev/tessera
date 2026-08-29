package user

import "testing"

func BenchmarkKeyHasher_Hash(b *testing.B) {
	h := NewKeyHasher()
	key := "tsr_v1_65NmxW3xXhhJoZYTdVY667pXK-qDomZ-e_PnVA1dFNE"

	for b.Loop() {
		_, _ = h.Hash(key)
	}
}

func BenchmarkKeyHasher_Verify(b *testing.B) {
	h := NewKeyHasher()
	key := "tsr_v1_65NmxW3xXhhJoZYTdVY667pXK-qDomZ-e_PnVA1dFNE"
	hash, _ := h.Hash(key)
	b.ResetTimer()
	for b.Loop() {
		_, _ = h.Verify(key, hash)
	}
}
