package task01

import "testing"

// TODO: напиши табличные тесты для функции ParsePrice.
//
// Создай срез структур с полями:
//   name        string   - название теста
//   input       string   - входная строка
//   expected    float64  - ожидаемый результат
//   expectError bool     - ожидаем ли ошибку
//
// Обязательно проверь эти случаи:
//   Корректные: "1500", "1 500", "1500.50", "1500,50", "1 500,50 руб.", "₽1 500"
//   Граничные:  "0", "0,01"
//   Некорректные: "", "abc", "руб.", "---"
//
// Подсказка: запусти go test -v ./... чтобы увидеть подробный вывод

func TestParsePrice(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		// TODO: заполни таблицу тест-кейсов
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: вызови ParsePrice(tc.input)
			// TODO: проверь что err != nil если expectError == true
			// TODO: проверь что результат == tc.expected (используй сравнение с погрешностью 0.001)
		})
	}
}

// FuzzParsePrice проверяет что функция не паникует ни при каких входных данных.
// Запуск: go test -fuzz=FuzzParsePrice -fuzztime=10s
func FuzzParsePrice(f *testing.F) {
	// TODO: добавь начальные случаи через f.Add(...)
	// f.Add("1500")
	// f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		// TODO: вызови ParsePrice(input)
		// Функция не должна паниковать - только возвращать ошибку
		// Не нужно проверять результат - главное что нет panic
	})
}
