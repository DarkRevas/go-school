package service

import (
	"context"
	"errors"

	"github.com/yourname/my-blog-service/internal/domain"
	"github.com/yourname/my-blog-service/internal/repository"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, name string, price float64, stock int) (domain.Product, error) {
	p := domain.Product{
		Name:  name,
		Price: price,
		Stock: stock,
	}
	return s.repo.Save(ctx, p)
}

func (s *ProductService) List(ctx context.Context) ([]domain.Product, error) {
	return s.repo.FindAll(ctx)
}

func (s *ProductService) Buy(ctx context.Context, id int, quantity int) (domain.Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}

	if p.Stock < quantity {
		return domain.Product{}, ErrInsufficientStock
	}

	p.Stock -= quantity
	return s.repo.Save(ctx, p)
}
