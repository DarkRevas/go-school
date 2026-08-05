package main

import (
	"strings"
	"sync"
	"testing"
)

func TestH2O(t *testing.T) {
	const molecules = 10
	const total = molecules * 3 // 10 * (2H + 1O) = 30 атомов

	h2o := NewH2O()
	var mu sync.Mutex
	var result strings.Builder
	var wg sync.WaitGroup

	for range molecules * 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h2o.Hydrogen(func() {
				mu.Lock()
				result.WriteRune('H')
				mu.Unlock()
			})
		}()
	}

	for range molecules {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h2o.Oxygen(func() {
				mu.Lock()
				result.WriteRune('O')
				mu.Unlock()
			})
		}()
	}

	wg.Wait()

	s := result.String()
	if len(s) != total {
		t.Fatalf("ожидали %d символов, получили %d", total, len(s))
	}

	hCount := strings.Count(s, "H")
	oCount := strings.Count(s, "O")
	if hCount != molecules*2 {
		t.Errorf("H count = %d, want %d", hCount, molecules*2)
	}
	if oCount != molecules {
		t.Errorf("O count = %d, want %d", oCount, molecules)
	}
}
