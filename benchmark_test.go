package main

import (
	"os"
	"testing"
)

func BenchmarkTipsData_addTip(b *testing.B) {
	td := &TipsData{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		td.addTip("benchmark", "This is a benchmark tip for performance testing")
	}
}

func BenchmarkTipsData_getRandomTip(b *testing.B) {
	td := &TipsData{}

	for i := 0; i < 1000; i++ {
		td.addTip("benchmark", "Benchmark tip content for performance testing")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		td.getRandomTip([]string{"benchmark"})
	}
}

func BenchmarkTipsData_removeTip(b *testing.B) {
	td := &TipsData{}
	tipIDs := make([]string, b.N)

	for i := 0; i < b.N; i++ {
		td.addTip("benchmark", "Benchmark tip content for removal testing")
		if len(td.Tips) > 0 {
			tipIDs[i] = td.Tips[len(td.Tips)-1].ID
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		td.removeTip(tipIDs[i])
	}
}

func BenchmarkLoadTips(b *testing.B) {
	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testData := &TipsData{}
	for i := 0; i < 100; i++ {
		testData.addTip("benchmark", "Benchmark tip content for load testing")
	}

	if err := saveTips(testData); err != nil {
		b.Fatalf("Failed to save test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loadTips()
		if err != nil {
			b.Fatalf("loadTips failed: %v", err)
		}
	}
}

func BenchmarkSaveTips(b *testing.B) {
	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testData := &TipsData{}
	for i := 0; i < 100; i++ {
		testData.addTip("benchmark", "Benchmark tip content for save testing")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := saveTips(testData)
		if err != nil {
			b.Fatalf("saveTips failed: %v", err)
		}
	}
}

func BenchmarkTipFiltering(b *testing.B) {
	td := &TipsData{}

	topics := []string{"git", "vim", "bash", "docker", "kubernetes"}
	for _, topic := range topics {
		for i := 0; i < 200; i++ {
			td.addTip(topic, "Benchmark tip content for filtering testing")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		td.getRandomTip([]string{"git", "docker"})
	}
}

func BenchmarkLargeDatasetOperations(b *testing.B) {
	td := &TipsData{}

	for i := 0; i < 10000; i++ {
		td.addTip("large-dataset", "This is tip content for large dataset benchmark testing")
	}

	b.Run("RandomSelection", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			td.getRandomTip([]string{"large-dataset"})
		}
	})

	b.Run("FilteredSelection", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			td.getRandomTip([]string{"large-dataset", "nonexistent"})
		}
	})
}

func BenchmarkConcurrentAccess(b *testing.B) {
	td := &TipsData{}

	for i := 0; i < 1000; i++ {
		td.addTip("concurrent", "Concurrent access benchmark tip")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			td.getRandomTip([]string{"concurrent"})
		}
	})
}

func BenchmarkJSONOperations(b *testing.B) {
	td := &TipsData{}

	for i := 0; i < 1000; i++ {
		td.addTip("json-bench", "This is a benchmark tip with realistic content length for JSON serialization testing")
	}

	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	b.Run("Save", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := saveTips(td)
			if err != nil {
				b.Fatalf("saveTips failed: %v", err)
			}
		}
	})

	if err := saveTips(td); err != nil {
		b.Fatalf("Failed to save data for load benchmark: %v", err)
	}

	b.Run("Load", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := loadTips()
			if err != nil {
				b.Fatalf("loadTips failed: %v", err)
			}
		}
	})
}
