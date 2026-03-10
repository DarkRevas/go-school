package service

import "github.com/go-course/clean-arch-task01/repository"

// TODO: создай структуру ProductService с полем repo типа repository.ProductRepository (интерфейс!)
// TODO: создай конструктор NewProductService(repo repository.ProductRepository) *ProductService

// TODO: реализуй методы:
//   Create(name string, price float64, stock int) (domain.Product, error)
//     - создаёт продукт через domain.Product, валидирует, сохраняет
//
//   List() ([]domain.Product, error)
//     - возвращает все продукты
//
//   Buy(productID int, quantity int) error
//     - находит продукт, проверяет что quantity <= Stock
//     - если ок - уменьшает Stock через UpdateStock
//     - если нет - возвращает ошибку "недостаточно товара на складе"

_ = repository.ProductRepository(nil) // убери эту строку когда начнёшь писать код
