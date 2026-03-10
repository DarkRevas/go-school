// Задание 2: Подсчёт гласных букв
//
// Напиши функцию countVowels(s string) int,
// которая принимает строку и возвращает количество гласных букв в ней.
// Учитывай и русские, и латинские гласные.
//
// Русские гласные: а, е, ё, и, й, о, у, ы, э, ю, я (и заглавные тоже)
// Латинские гласные: a, e, i, o, u (и заглавные тоже)
//
// Подсказка: используй for range по строке - он автоматически перебирает
// символы (rune), а не байты. Это важно для кириллицы!
//
// Ожидаемый вывод:
//   "Привет мир" -> гласных: 4
//   "Hello World" -> гласных: 3
//   "Go" -> гласных: 1
//
// Запусти: go run main.go

package main

import (
	"fmt"
	"unicode"
)

// TODO: напиши функцию countVowels(s string) int
// Внутри используй for range и switch для проверки каждого символа
func countVowels(s string) int {
	count := 0
	for _, r := range s {
		switch unicode.ToLower(r) {
		// Русские гласные
		case 'а', 'е', 'ё', 'и', 'й', 'о', 'у', 'ы', 'э', 'ю', 'я':
			count++
		// Латинские гласные
		case 'a', 'e', 'i', 'o', 'u':
			count++
		}
	}
	return count
}


func main() {
	tests := []string{"Привет мир", "Hello World", "Go"}
	for _, s := range tests {
		fmt.Printf("%q -> гласных: %d\n", s, countVowels(s))
	}
}
