package main

import "fmt"

type Book struct {
	Title string
	Pages int
}

func (b Book) Summary() string {
	return fmt.Sprintf("%s (%d pages)", b.Title, b.Pages)
}

type Wallet struct {
	Balance int
}

func (w *Wallet) Deposit(amount int) {
	w.Balance += amount
}

func (w Wallet) DepositCopy(amount int) {
	w.Balance += amount
	fmt.Printf("DepositCopy inside func balance: %v \n", w.Balance)
}

func resetScore(score *int) {
	*score = 0
}

type User struct {
	Name string
}

func printUserName(u *User) {
	if u == nil {
		fmt.Println("user is nil")
		return
	}
	fmt.Println(u.Name)
}

// Создай структуру Timer:
	// Seconds int
	// Running bool

	// Добавь методы:

	// Start() (указатель) — ставит Running = true;
	// Stop() (указатель) — ставит Running = false;
	// Status() string (значение) — возвращает "running" или "stopped".
type Timer struct {
	Seconds int
	Running bool
}

func (t *Timer) Start() {
	if t == nil {
		return
	}
	t.Running = true
}

func (t *Timer) Stop() {
	if t == nil {
		return
	}
	t.Running = false
}

func (t Timer) Status() string {
	if t.Running {
		return "running"
	}
	return "stopped"
}

type BankAccount struct {
	Owner   string
	Balance int
}

func (b *BankAccount) Deposit(amount int) {
	if amount > 0 {
		b.Balance += amount
	}
}

func (b *BankAccount) Withdraw(amount int) bool {
	if amount > b.Balance || amount <= 0 {
		return false
	}

	b.Balance -= amount
	return true
}

func (b BankAccount) Info() string {
	return fmt.Sprintf("Owner: %s, Balance: %d", b.Owner, b.Balance)
}

func main() {
	// Задача 3.2.1. Метод чтения

	// Создай структуру Book с полями:
	// Title string
	// Pages int

	// Добавь метод Summary() string с получателем-значением, который возвращает строку: "<Title> (<Pages> pages)".

	book := Book{
		Title: "Go Basics",
		Pages: 250,
	}

	fmt.Println(book.Summary()) // Go Basics (250 pages)

	// Задача 3.2.2. Изменение через указатель

	// Создай структуру Wallet с полем:
	// Balance int

	// Добавь метод Deposit(amount int) с получателем-указателем, который увеличивает баланс.
	// В main сделай два пополнения и выведи итоговый баланс.

	wallet := Wallet{
		Balance: 100,
	}

	wallet.Deposit(899)
	wallet.Deposit(1)
	fmt.Println(wallet) // {1000}

	// Задача 3.2.3. Сравнение двух подходов

	// Для структуры Wallet добавь:
	// DepositCopy(amount int) с получателем-значением;
	// Deposit(amount int) с получателем-указателем.

	// Покажи в main, что первый не меняет исходный объект, а второй меняет.

	wallet.DepositCopy(899) // Inside DepositCopy 1899
	wallet.DepositCopy(1)   // Inside DepositCopy 1001
	fmt.Printf("DepositCopy outside func balance: %v \n", wallet) // {1000}

	// Задача 3.2.4. Функция с указателем на базовый тип

	// Напиши функцию resetScore(score *int), которая устанавливает значение в 0.
	// В main создай переменную score, вызови функцию и выведи результат.

	score := 999
	resetScore(&score)
	fmt.Printf("Score: %v \n", score) // Score: 0

	// Задача 3.2.5. Безопасная работа с nil

	// Создай структуру User с полем Name.
	// Напиши функцию printUserName(u *User), которая:
	// печатает user is nil, если u == nil;
	// иначе печатает имя пользователя.
	// Проверь обе ветки в main.
	var user *User
	printUserName(user) // user is nil
	user = &User{
		Name: "Anton",
	}
	printUserName(user) // Anton
	// Задача 3.2.6. Небольшая модель состояния

	// Создай структуру Timer:
	// Seconds int
	// Running bool

	// Добавь методы:

	// Start() (указатель) — ставит Running = true;
	// Stop() (указатель) — ставит Running = false;
	// Status() string (значение) — возвращает "running" или "stopped".
	timer := Timer{
		Seconds: 100,
		Running: false,
	}
	fmt.Println(timer.Status()) // stopped
	timer.Start()
	fmt.Println(timer.Status()) // running
	timer.Stop()
	fmt.Println(timer.Status()) // stopped

	// Задача 3.2.7. Интеграционная задача

	// Спроектируй структуру BankAccount:
	// Owner string
	// Balance int

	// И реализуй методы:

	// Deposit(amount int) (указатель),
	// Withdraw(amount int) bool (указатель, возвращает false, если недостаточно средств),
	// Info() string (значение).
	// В main создай счет, выполни несколько операций и выведи финальное состояние.
	
	account := BankAccount{
		Owner:   "Alice",
		Balance: 100,
	}

	fmt.Println(account.Info()) //Owner: Alice, Balance: 100

	account.Deposit(50)
	fmt.Println(account.Info()) // Owner: Alice, Balance: 150

	ok := account.Withdraw(120)
	fmt.Println("Withdraw success:", ok) // Withdraw success: true
	fmt.Println(account.Info()) // Owner: Alice, Balance: 30

	notOk := account.Withdraw(120) 
	fmt.Println("Withdraw not success:", notOk) // Withdraw not success: false
	fmt.Println(account.Info()) // Owner: Alice, Balance: 30
}