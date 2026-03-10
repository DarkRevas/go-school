package repository

import "github.com/go-course/clean-arch-task01/domain"

// TODO: создай интерфейс ProductRepository с методами:
//   Save(product domain.Product) (domain.Product, error)
//   FindAll() ([]domain.Product, error)
//   FindByID(id int) (domain.Product, error)
//   Delete(id int) error
//   UpdateStock(id int, newStock int) error

// TODO: создай структуру InMemoryProductRepository и реализуй все методы интерфейса
// Используй map[int]domain.Product для хранения и sync.RWMutex для защиты от гонок

// TODO: создай конструктор NewInMemoryProductRepository() *InMemoryProductRepository

_ = domain.Product{} // убери эту строку когда начнёшь писать код
