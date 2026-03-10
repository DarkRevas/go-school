// Задание 2: Идентификатор запроса через контекст
//
// Нужно передать идентификатор запроса через цепочку функций с помощью context.WithValue.
//
// 1. Создай свой тип для ключа контекста:
//    type contextKey string
//    const requestIDKey contextKey = "request-id"
//
// 2. Напиши функцию middleware(next func(ctx context.Context)):
//    - создаёт контекст с идентификатором "req-42" через context.WithValue
//    - вызывает next(ctx) передавая обогащённый контекст
//
// 3. Напиши функцию handler(ctx context.Context):
//    - достаёт идентификатор из контекста через ctx.Value(requestIDKey)
//    - выводит: "Обрабатываем запрос: req-42"
//    - если идентификатора нет - выводит предупреждение
//
// Ожидаемый вывод:
//   Обрабатываем запрос: req-42
//
// Запусти: go run main.go

package main

import (
	"context"
	"fmt"
)

// TODO: объяви тип ключа и константу

// TODO: напиши функцию middleware(next func(ctx context.Context))

// TODO: напиши функцию handler(ctx context.Context)

func main() {
	middleware(handler)

	_ = context.Background() // убери когда начнёшь использовать
	_ = fmt.Println          // убери когда начнёшь использовать
}
