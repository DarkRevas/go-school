package main

import (
	"errors"
	"fmt"
)

type Greeter interface {
	Greet() string
}

type User struct{}
func (u User) Greet() string {
	return "im User"
} 
type Guest struct{}
func (g Guest) Greet() string {
	return "im Guest"
}

type Shape interface {
	Area() float64
}

type Rectangle struct {
	Width float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

type Square struct {
	Side float64
}

func (s Square) Area() float64 {
	return s.Side * s.Side
}

type Runner interface {
	Run() string
}

type Athlete struct{}
func (a Athlete) Run() string {
	return "Athlete is running"
}

type Starter interface {
	Start()
}

type Stopper interface {
	Stop()
}

type Machine interface {
	Starter
	Stopper
}

type Engine struct {
	Name string
}

func (e Engine) Start() {
	fmt.Println(e.Name, "started")
}

func (e Engine) Stop() {
	fmt.Println(e.Name, "stopped")
}

func runCycle(m Machine) {
	m.Start()
	fmt.Println("Machine is running...")
	m.Stop()
}

func printStringLength(x any) {
	str, ok := x.(string)
	if !ok {
		fmt.Println("not a string")
		return
	}
	fmt.Println(len(str))
}

type PaymentProcessor interface {
	Pay(amount int) error
}

type CardProcessor struct{}
func (c CardProcessor) Pay(amount int) error {
	if amount < 0 {
		return errors.New("Invalid amount")
	}
	fmt.Printf("CardProcessor paid: %v", amount)
	return nil
}

type CashProcessor struct{}
func (c CashProcessor) Pay(amount int) error{
	if amount < 0 {
		return errors.New("Invalid amount")
	}
	fmt.Printf("CashProcessor paid: %v", amount)
	return nil
}

func checkout(p PaymentProcessor, amount int) {
	err := p.Pay(amount)
	if err != nil {
		fmt.Println("Payment failed:", err)
		return
	}

	fmt.Println("Checkout successful")
}

// Спроектируй систему логирования с интерфейсом Logger:
// Log(message string)
// Сделай две реализации:
// ConsoleLogger
// PrefixLogger (имеет поле Prefix и добавляет его к сообщению)

// Создай функцию processOrder(logger Logger, id string), которая пишет 2-3 сообщения о процессе заказа.

type Logger interface {
	Log(message string)
}

func processOrder(logger Logger, id string) {
	logger.Log("Start processing order " + id)
	logger.Log("Checking inventory for order " + id)
	logger.Log("Order " + id + " completed")
}

type ConsoleLogger struct {}

func (c ConsoleLogger) Log(message string) {
	fmt.Println(message)
}

type PrefixLogger struct {
	Prefix string
}

func (p PrefixLogger) Log(message string) {
	fmt.Println(p.Prefix + " " + message)
}

func main() {
// Задача 3.3.1. Базовый интерфейс

// Создай интерфейс Greeter с методом Greet() string.
// Сделай два типа (User, Guest), каждый возвращает свое приветствие.
// Напиши функцию printGreeting(g Greeter).
user := User{}
fmt.Println(user.Greet()) // im User
guest := Guest{}
fmt.Println(guest.Greet()) // im Guest

// Задача 3.3.2. Срез интерфейсов

// Создай интерфейс Shape с методом Area() float64.
// Реализуй его для:
// Rectangle (Width, Height)
// Square (Side)

// Сохрани фигуры в []Shape и выведи площадь каждой.
shape := []Shape{Rectangle{10,2}, Square{5}}
for i, el := range shape {
	fmt.Printf("index: %v, Area: %v \n",i ,el.Area()) // index: 0, Area: 20 // index: 1, Area: 25 
}

// Задача 3.3.3. Проверка реализации через компилятор

// Создай интерфейс Runner с методом Run() string.
// Определи тип Athlete, который реализует Run().
// Добавь строку:
// var _ Runner = Athlete{}
// Убедись, что код компилируется.

var _ Runner = Athlete{} // ok

// Задача 3.3.4. Композиция интерфейсов

// Создай:
// Starter с методом Start()
// Stopper с методом Stop()
// Machine как композицию Starter и Stopper
// Сделай тип Engine, который реализует оба метода.
// Напиши функцию runCycle(m Machine).

engine := Engine{Name: "V8 Engine"}

runCycle(engine)

// Задача 3.3.5. Безопасный type assertion

// Напиши функцию printStringLength(x any), которая:
// пытается привести x к string,
// если успешно, печатает длину строки,
// если нет, печатает not a string.
itsStr := "string"
printStringLength(itsStr) // 6
notString := false
printStringLength(notString) // not a string

// Задача 3.3.6. Интерфейс для обработки платежа

// Создай интерфейс PaymentProcessor с методом Pay(amount int) error.
// Реализуй две структуры:
// CardProcessor
// CashProcessor
// Сделай функцию checkout(p PaymentProcessor, amount int).
card := CardProcessor{}
cash := CashProcessor{}

checkout(card, 100) // CardProcessor paid: 100 Checkout successful
checkout(cash, 50) // CashProcessor paid: 50 Checkout successful

// Задача 3.3.7. Интеграционная задача

// Спроектируй систему логирования с интерфейсом Logger:
// Log(message string)
// Сделай две реализации:
// ConsoleLogger
// PrefixLogger (имеет поле Prefix и добавляет его к сообщению)

// Создай функцию processOrder(logger Logger, id string), которая пишет 2-3 сообщения о процессе заказа.
prefix := PrefixLogger{
	Prefix: "Prefix",
}
console := ConsoleLogger{}
processOrder(prefix, "1")
processOrder(console, "2")
}