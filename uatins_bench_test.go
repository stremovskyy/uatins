package uatins

import "testing"

func BenchmarkValidate(b *testing.B) {
	// Use a fixed, valid TIN synthesized as in tests
	const tin = "32874123450" // placeholder; replace with a valid synthesized one if needed
	c := NewClient(WithStrict(false))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Validate(tin, nil)
	}
}
