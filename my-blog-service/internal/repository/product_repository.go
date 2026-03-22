package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/yourname/my-blog-service/internal/domain"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository interface {
	Save(ctx context.Context, p domain.Product) (domain.Product, error)
	FindAll(ctx context.Context) ([]domain.Product, error)
	FindByID(ctx context.Context, id int) (domain.Product, error)
	Delete(ctx context.Context, id int) error
}

type InMemoryProductRepository struct {
	mu       sync.RWMutex
	products map[int]domain.Product
	nextID   int
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		products: make(map[int]domain.Product),
		nextID:   1,
	}
}

func (r *InMemoryProductRepository) Save(ctx context.Context, p domain.Product) (domain.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p.ID == 0 {
		p.ID = r.nextID
		r.nextID++
	}

	r.products[p.ID] = p
	return p, nil
}

func (r *InMemoryProductRepository) FindAll(ctx context.Context) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Product, 0, len(r.products))
	for _, p := range r.products {
		result = append(result, p)
	}
	return result, nil
}

func (r *InMemoryProductRepository) FindByID(ctx context.Context, id int) (domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.products[id]
	if !exists {
		return domain.Product{}, ErrProductNotFound
	}
	return p, nil
}

func (r *InMemoryProductRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[id]; !exists {
		return ErrProductNotFound
	}

	delete(r.products, id)
	return nil
}
