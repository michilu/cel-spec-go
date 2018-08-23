package example

import "testing"

func Benchmark_example(b *testing.B) {
	for i := 0; i < b.N; i++ {
		example("", "", "")
	}
}

func Benchmark_PaypalmeFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PaypalmeFunc("", "", "")
	}
}
