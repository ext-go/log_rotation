package log_rotation

import "testing"

func BenchmarkQueue(b *testing.B) {
	queue := newChan()
	data := []byte("test")
	b.ResetTimer()
	b.Run("insert", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			queue.put(data)
		}
	})
	b.Run("get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = queue.get()
		}
	})
}
