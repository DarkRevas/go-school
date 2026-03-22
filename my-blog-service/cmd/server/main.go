package main

import (
	"context"
	"fmt"
	"log"

	"go.uber.org/fx"

	"github.com/yourname/my-blog-service/internal/repository"
	"github.com/yourname/my-blog-service/internal/service"
)

func main() {
	fx.New(
		// Provide dependencies
		fx.Provide(
			fx.Annotate(
				repository.NewInMemoryProductRepository,
				fx.As(new(repository.ProductRepository)),
			),
			service.NewProductService,
		),
		// Invoke application logic
		fx.Invoke(run),
	).Run()
}

func run(lc fx.Lifecycle, svc *service.ProductService) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Сервер запущен")

			// Demo: Create products
			p1, err := svc.Create(ctx, "Laptop", 999.99, 10)
			if err != nil {
				return err
			}
			log.Printf("Created product: %+v", p1)

			p2, err := svc.Create(ctx, "Mouse", 29.99, 50)
			if err != nil {
				return err
			}
			log.Printf("Created product: %+v", p2)

			// Demo: Buy product
			updated, err := svc.Buy(ctx, p1.ID, 2)
			if err != nil {
				return err
			}
			log.Printf("Bought 2 items, remaining stock: %d", updated.Stock)

			// Demo: List all products
			products, err := svc.List(ctx)
			if err != nil {
				return err
			}
			log.Printf("All products: %d", len(products))
			for _, p := range products {
				fmt.Printf("  - %s: $%.2f (stock: %d)\n", p.Name, p.Price, p.Stock)
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Сервер остановлен")
			return nil
		},
	})
}
