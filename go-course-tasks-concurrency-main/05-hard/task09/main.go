// ============================================================
// Задача: Parallel ForEach с лимитом параллелизма  🟡 Middle
// ============================================================
//
// Часто на собесах: "у тебя slice из 10000 элементов и API с лимитом 20
// одновременных запросов — напиши обёртку".
//
// Реализуй:
//
//   func ParallelForEach[T any](
//       ctx context.Context,
//       items []T,
//       parallelism int,
//       fn func(ctx context.Context, item T) error,
//   ) error
//
// Требования:
//   - Обрабатывается максимум parallelism элементов одновременно
//   - Если fn возвращает ошибку — отменяется ctx производный, остальные fn
//     видят это в своём ctx.Done() и могут завершиться раньше
//   - Возвращается ПЕРВАЯ полученная ошибка
//   - Если ctx на входе отменён — возвращается ctx.Err()
//   - Нет утечек горутин
//
// Бонус:
//   func ParallelMap[I, O any](
//       ctx context.Context,
//       items []I,
//       parallelism int,
//       fn func(ctx context.Context, item I) (O, error),
//   ) ([]O, error)
//
//   Результаты в том же порядке что входы (см. 01-channels/task07).
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TODO: реализуй ParallelForEach
// Подсказка: семафор-канал ёмкости parallelism + context.WithCancel
// Можно собрать поверх errgroup из 03-patterns/task06 — но здесь
// постарайся реализовать напрямую чтобы понять механику.
func ParallelForEach[T any](
	ctx context.Context,
	items []T,
	parallelism int,
	fn func(ctx context.Context, item T) error,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if parallelism < 1 {
		parallelism = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(item T) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := fn(ctx, item); err != nil {
				once.Do(func() {
					firstErr = err
					cancel()
				})
			}
		}(item)
	}
	wg.Wait()
	return firstErr
}

// TODO: реализуй ParallelMap (бонус)
func ParallelMap[I, O any](
	ctx context.Context,
	items []I,
	parallelism int,
	fn func(ctx context.Context, item I) (O, error),
) ([]O, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if parallelism < 1 {
		parallelism = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make([]O, len(items))
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	for i, item := range items {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(i int, item I) {
			defer wg.Done()
			defer func() { <-sem }()
			v, err := fn(ctx, item)
			if err != nil {
				once.Do(func() {
					firstErr = err
					cancel()
				})
				return
			}
			out[i] = v
		}(i, item)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}

func main() {
	ctx := context.Background()
	items := make([]int, 50)
	for i := range items {
		items[i] = i
	}

	var done atomic.Int32
	err := ParallelForEach(ctx, items, 5, func(ctx context.Context, n int) error {
		select {
		case <-time.After(10 * time.Millisecond):
			done.Add(1)
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	fmt.Printf("done=%d err=%v\n", done.Load(), err) // done=50 err=nil

	// ParallelMap — квадраты
	squares, err := ParallelMap(ctx, []int{1, 2, 3, 4, 5}, 3,
		func(ctx context.Context, n int) (int, error) { return n * n, nil })
	fmt.Printf("squares=%v err=%v\n", squares, err) // [1 4 9 16 25]
}
