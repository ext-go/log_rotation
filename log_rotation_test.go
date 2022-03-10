package log_rotation

import (
	"log"
	"testing"
	"time"
)

func TestLogRotation(t *testing.T) {
	w := NewLogRotation(SetFileName("./test.log"), SetLimitUseMaxLines(1), EnableStdOut())
	w.Launch()
	log.SetOutput(w)
	log.Println("test 1")
	log.Println("test 2")
	time.Sleep(time.Second * 3)
}

func BenchmarkNewLogRotation(b *testing.B) {
	w := NewLogRotation(SetFileName("./test_bench.log"), SetLimitUseFileSize(1024))
	w.Launch()
	log.SetOutput(w)
	b.ResetTimer()
	b.Run("log", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			log.Println("test", i)
		}
	})
}
