// ============================================================
// Задача: Timeout & Select — первый ответ выигрывает  🟡 Middle
// ============================================================
//
// Реализуй функцию fastest(ctx, urls) которая:
//   - Параллельно запрашивает все переданные URL (mock)
//   - Возвращает первый успешный ответ
//   - Отменяет остальные запросы
//   - Возвращает ошибку если все запросы упали или истёк таймаут ctx
//
// Это классический паттерн "hedged requests" широко используемый в продакшне.
//
// Дополнительно реализуй withTimeout(d, fn):
//   - Запускает fn с таймаутом d
//   - Если fn не завершилась за d — возвращает ErrTimeout
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var ErrTimeout = errors.New("таймаут истёк")
var ErrAllFailed = errors.New("все запросы завершились ошибкой")

type Result struct {
	URL  string
	Body string
}

// mockFetch имитирует HTTP-запрос с случайной задержкой
func mockFetch(ctx context.Context, url string) (Result, error) {
	delay := time.Duration(50+rand.Intn(200)) * time.Millisecond

	select {
	case <-time.After(delay):
		if rand.Float64() < 0.2 { // 20% вероятность ошибки
			return Result{}, fmt.Errorf("%s: server error", url)
		}
		return Result{URL: url, Body: fmt.Sprintf("response from %s", url)}, nil
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

// TODO: реализуй fastest
// Подсказка: отмена ctx распространяется на все запросы автоматически — не заботься об явном завершении
func fastest(ctx context.Context, urls []string) (Result, error) {
	ctxInside, cancel := context.WithCancel(ctx)
	defer cancel()
	// TODO: реализуй
	type response struct {
		resp Result
		err error
	}

	ch := make(chan response)
	wg := sync.WaitGroup{}
	for _, url := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := mockFetch(ctxInside, url)
			select {
			case <-ctxInside.Done():
				return
			case ch <- response{resp: res, err: err}:
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var lastError error

	for el := range ch {
		if el.err == nil {
			cancel()
			return Result{URL: el.resp.URL, Body: el.resp.Body}, nil
		}
		lastError = el.err
	}

	if lastError != nil {
		return Result{}, ErrAllFailed
	}

	return Result{}, ctx.Err()
}

// TODO: реализуй withTimeout
func withTimeout(d time.Duration, fn func() (string, error)) (string, error) {
	// TODO
	type response struct {
		resp string
		err error
	}
	ch := make(chan response, 1)

	go func() {
		res, err := fn()
		ch <- response{resp: res, err: err}
	}()

	select {
	case <-time.After(d):
		return "", ErrTimeout
	case el := <-ch:
		return el.resp, el.err
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	urls := []string{
		"https://api1.example.com",
		"https://api2.example.com",
		"https://api3.example.com",
	}

	result, err := fastest(ctx, urls)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Printf("Быстрейший ответ от %s: %s\n", result.URL, result.Body)
}
