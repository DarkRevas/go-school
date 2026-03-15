package main

import (
	"errors"
	"fmt"
	"strconv"
)

func safeDivide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("Нельзя делить на 0")
	}
	return a/b, nil
}

func createUser(name string) error {
	if name == "" {
		return errors.New("Name is empty")
	}
	fmt.Println("User name is " + name)
	return nil
}

var ErrOutOfStock = errors.New("out of stock")

func buyItem(count int) error {
	if count == 0 {
		return ErrOutOfStock
	}
	fmt.Printf("Bought: %v items \n", count)
	return nil
}

func readFileMock() error {
	return errors.New("read file mock error")
}

func loadData() error {
	err := readFileMock()
	if err != nil {
		return fmt.Errorf("wrap error loadData: %w", err)
	}
	return nil
}

type InputError struct {
	Field string
	Reason string
}

func (i InputError) Error() string {
	return fmt.Sprintf("Reason: %s, Field: %s", i.Reason, i.Field)
}

func validateEmail(email string) error {
	if email == "" {
		return InputError{Field: "email", Reason: "empty"}
	}

	return nil
}

func parseID(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty ID")
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid ID '%s': %w", s, err)
	}
	return id, nil
}

func validateName(name string) error {
	if name == "" {
		return errors.New("имя не может быть пустым")
	}
	return nil
}

func validateAge(age int) error {
	if age < 18 {
		return fmt.Errorf("возраст слишком мал: %d", age)
	}
	if age > 100 {
		return fmt.Errorf("возраст слишком велик: %d", age)
	}
	return nil
}

func register(name string, age int) (string, error) {
	if err := validateName(name); err != nil {
		return "", fmt.Errorf("ошибка в имени: %w", err)
	}
	if err := validateAge(age); err != nil {
		return "", fmt.Errorf("ошибка в возрасте: %w", err)
	}
	return fmt.Sprintf("зарегистрировано имя: %s, возраст: %v", name, age),nil
}

func main() {
// Задача 3.4.1. Деление с проверкой

// Напиши функцию safeDivide(a, b int) (int, error).
// Если b == 0, верни ошибку.
// В main вызови функцию с корректным и некорректным входом.
{
	res, err := safeDivide(10,0)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}

{
	res, err := safeDivide(10,5)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}

// Задача 3.4.2. Проверка пустого имени

// Напиши функцию createUser(name string) error.
// Если имя пустое, верни ошибку errors.New("name is empty").
// В main проверь оба сценария.
{
	err := createUser("")
	if err != nil {
		fmt.Println(err)
	}
}

{
	err := createUser("Goga")
	if err != nil {
		fmt.Println(err)
	}
}

// Задача 3.4.3. Sentinel error

// Объяви:
// var ErrOutOfStock = errors.New("out of stock")
// Напиши функцию buyItem(count int) error, которая возвращает ErrOutOfStock, если count == 0.
// В main обработай этот случай через errors.Is.
{
	err := buyItem(0)
	if err != nil {
		if errors.Is(err, ErrOutOfStock) {
			fmt.Println(ErrOutOfStock)
		} else {
			fmt.Println("Unknown error:", err)
		}
	}
}

{
	err := buyItem(10)
	if err != nil {
		if errors.Is(err, ErrOutOfStock) {
			fmt.Println(ErrOutOfStock)
		} else {
			fmt.Println("Unknown error:", err)
		}
	}
}


// Задача 3.4.4. Оборачивание ошибки

// Сделай две функции:
// readFileMock() error — возвращает базовую ошибку;
// loadData() error — вызывает первую и оборачивает ошибку через %w.
// В main выведи итоговую ошибку.
{
	err := loadData()
	if err != nil {
		fmt.Println(err)
	}
}

// Задача 3.4.5. Кастомная ошибка

// Создай тип InputError с полями:
// Field string
// Reason string
// Реализуй метод Error() string.
// Напиши функцию validateEmail(email string) error, которая возвращает InputError, если строка пустая.
// В main извлеки ошибку через errors.As.
{
	err := validateEmail("")
	if err != nil {
		var inputErr InputError
		if errors.As(err, &inputErr) {
			fmt.Println("Unwrapped error:", inputErr)
		} else {
			fmt.Println("Unknown error:", err)
		}
	}
}
// Задача 3.4.6. Обработка ошибок в цикле

// Дан срез строковых значений IDs.
// Напиши функцию parseID(s string) (int, error), которая возвращает ошибку для пустой строки.
// В цикле:
// корректные значения печатай как ok;
// ошибочные — как skip и переходи к следующему элементу.
{
	sl := []string{"1", "0", "", "5", "", "3", ""}
	for _, el := range sl {
		_, err := parseID(el)
		if err != nil {
			fmt.Println("skip")
			continue
		}
		fmt.Println("ok")
	}
	// ok
	// ok
	// skip
	// ok
	// skip
	// ok
	// skip
}


// Задача 3.4.7. Интеграционная задача

// Спроектируй мини-поток регистрации пользователя:

// validateName(name string) error
// validateAge(age int) error
// register(name string, age int) error

// Требования:
// register вызывает обе проверки;
// ошибки оборачиваются с контекстом;
// в main протестируй минимум 3 сценария (успех и два разных отказа).
{
	result, err := register("Абобус", 10)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)

	result, err = register("", 20)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)

	result, err = register("Пушкин", 30)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}
}