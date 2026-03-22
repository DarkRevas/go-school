package task01

import (
	"math"
	"testing"
)

func TestParsePrice(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		// Корректные случаи
		{"simple integer", "1500", 1500.00, false},
		{"with space separator", "1 500", 1500.00, false},
		{"with dot decimal", "1500.50", 1500.50, false},
		{"with comma decimal", "1500,50", 1500.50, false},
		{"space and comma", "1 500,50", 1500.50, false},
		{"with rub suffix", "1 500,50 руб", 1500.50, false},
		{"with ruble symbol prefix", "₽1 500", 1500.00, false},
		{"with multiple spaces", "1  500  ,  50", 1500.50, false},
		{"large number", "1 000 000,99", 1000000.99, false},

		// Граничные случаи
		{"zero", "0", 0.00, false},
		{"minimal positive", "0,01", 0.01, false},
		{"minimal with dot", "0.01", 0.01, false},
		{"whitespace around", "  1500  ", 1500.00, false},

		// Некорректные случаи
		{"empty string", "", 0, true},
		{"only letters", "abc", 0, true},
		{"only currency word", "руб.", 0, true},
		{"only dashes", "---", 0, true},
		{"only spaces", "   ", 0, true},
		{"only currency symbol", "₽", 0, true},
		{"multiple dots", "1.500.50", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParsePrice(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("ожидалась ошибка, но получено значение: %f", result)
				}
				return
			}

			if err != nil {
				t.Errorf("не ожидалась ошибка, но получена: %v", err)
				return
			}

			if math.Abs(result-tc.expected) > 0.001 {
				t.Errorf("ожидалось %f, получено %f", tc.expected, result)
			}
		})
	}
}

// FuzzParsePrice проверяет что функция не паникует ни при каких входных данных.
// Запуск: go test -fuzz=FuzzParsePrice -fuzztime=10s
func FuzzParsePrice(f *testing.F) {
	// Seed corpus with test cases
	f.Add("1500")
	f.Add("1 500")
	f.Add("1500.50")
	f.Add("1500,50")
	f.Add("1 500,50 руб")
	f.Add("₽1 500")
	f.Add("")
	f.Add("abc")
	f.Add("0")
	f.Add("0,01")
	f.Add("   ")
	f.Add("руб.")
	f.Add("---")
	f.Add("1 000 000,99")

	f.Fuzz(func(t *testing.T, input string) {
		// Функция не должна паниковать - только возвращать ошибку или значение
		_, _ = ParsePrice(input)
	})
}
