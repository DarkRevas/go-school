// Задание 2: Рабочее пространство (go work)
//
// Что нужно сделать:
//
// Шаг 1. Перейди в папку task02-workspace (не в module-a, а на уровень выше):
//   cd ..
//
// Шаг 2. Создай файл рабочего пространства командой:
//   go work init
//
// Шаг 3. Подключи оба модуля к рабочему пространству:
//   go work use ./module-a ./module-b
//
// Шаг 4. Перейди в module-b и создай файл hello.go с функцией SayHello.
//   Посмотри на TODO в файле module-b/hello.go.
//
// Шаг 5. Вернись в module-a, допиши вызов функции SayHello ниже.
//
// Шаг 6. Запусти из папки module-a:
//   go run main.go
//
// Ожидаемый результат: программа напечатает сообщение из module-b.

package main

import (
	// TODO: импортируй пакет из module-b
	// "github.com/yourname/module-b"
	"fmt"
)

func main() {
	// TODO: вызови функцию SayHello из пакета module-b
	// Пример: fmt.Println(moduleb.SayHello())

	fmt.Println("Допиши этот код!")
}
