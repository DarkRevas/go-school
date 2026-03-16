package main

import "fmt"

type Profile struct {
	Name string
	Age int
	IsActive bool
}

type AppConfig struct {
	Host string
	Port int
	Debug bool
}

type Address struct {
	City string
	Street string
}

type Employee struct {
	Name string
	Address Address
}

type Package struct {
	ID string
	Weight int
}

type Destination struct {
	City string
	Zip string
}

type Shipment struct {
	Package Package
	Destination Destination
}

type Audit struct {
	CreatedAt string
	UpdatedAt string
}

type Article struct {
	Title string
	Audit
}

type ContactInfo struct {
	Phone string
	Email string
}

type Client struct {
	ID string
	Address Address
	ContactInfo
}

type Student struct {
	ID   int
	Name string
	ContactInfo
}

type Course struct {
	ID    int
	Title string
	Hours int
}

type CourseEnrollment struct {
	EnrollmentID int
	Status       string

	Student // embedding
	Course  // embedding
}

func main() {
	// Задача 3.1.1 Профиль пользователя
	// Создай структуру Profile с полями:

	// Name string
	// Age int
	// IsActive bool

	// В main создай значение структуры через именованный литерал и выведи его поля.
	p := Profile{
		Name: "Tarantino",
		Age: 55,
		IsActive: true,
	}

	fmt.Printf("Profile Name: %s, Age: %v, isActive: %v \n", p.Name, p.Age, p.IsActive)

	// Задача 3.1.2. Нулевые значения

	// Создай структуру AppConfig с полями:

	// Host string
	// Port int
	// Debug bool

	// Объяви переменную var cfg AppConfig без инициализации.
	// Выведи всю структуру и отдельно каждое поле.
	// Определи, какие нулевые значения были установлены.

	var cfg AppConfig
	fmt.Println(cfg) // { 0 false}
	fmt.Println(cfg.Debug, cfg.Host, cfg.Port) // false  0
	fmt.Printf("%+v\n", cfg) // {Host: Port:0 Debug:false}

	// Задача 3.1.3. Вложенный адрес

	// Создай структуры:
	// Address (City, Street)
	// Employee (Name, Address Address)

	// Создай одного сотрудника и выведи строку:
	// <Name>: <City>, <Street>.

	e := Employee{
		Name: "Josh",
		Address: Address{
			Street: "Lomonosova st.",
			City: "Moscow",
		},
	}

	fmt.Printf("Name: %s, City: %s, Street: %s \n", e.Name, e.Address.City, e.Address.Street) // Name: Josh, City: Moscow, Street: Lomonosova st. 

	// Задача 3.1.4. Модель доставки

	// Создай структуры:
	// Package (ID string, Weight int)
	// Destination (City string, Zip string)
	// Shipment (Package Package, Destination Destination)
	// Создай значение Shipment и выведи ID посылки и город назначения.

	s := Shipment{
		Package: Package{
			ID: "123",
			Weight: 100,
		},
		Destination: Destination{
			City: "New York",
			Zip: "123456",
		},
	}

	fmt.Printf("ID: %s, City: %s \n", s.Package.ID, s.Destination.City) // ID: 123, City: New York

	// Задача 3.1.5. Композиция с Audit

	// Создай структуру Audit с полями:
	// CreatedAt string
	// UpdatedAt string

	// Создай структуру Article с полями:
	// Title string
	// встроенное поле Audit (embedding)
	// Инициализируй Article и выведи Title, CreatedAt, UpdatedAt.

	a := Article{
		Title: "Gaben",
		Audit: Audit{
			CreatedAt: "1999-02-01",
			UpdatedAt: "2000-03-01",
		},
	}

	fmt.Printf("Title: %s, CreatedAt: %s, UpdatedAt: %s \n", a.Title, a.CreatedAt, a.UpdatedAt) // Title: Gaben, CreatedAt: 1999-02-01, UpdatedAt: 2000-03-01 

	// Задача 3.1.6. Две переиспользуемые части

	// Создай структуры:
	// ContactInfo (Phone, Email)
	// Address (City, Street)
	// Client (ID string, Address Address, встроенное ContactInfo)

	// В main создай клиента и выведи:
	// ID
	// City
	// Email (доступ напрямую через embedding).

	cl := Client{
		ID: "123",
		Address: Address{
			City: "Sopot",
		},
		ContactInfo: ContactInfo{
			Email: "krasnoe@beloe.ru",
		},
	}

	fmt.Printf("ID: %s, City: %s, Email: %s \n", cl.ID, cl.Address.City, cl.Email) //ID: 123, City: Sopot, Email: krasnoe@beloe.ru 

	// Задача 3.1.7. Интеграционная мини-модель

	// Спроектируй и реализуй модель CourseEnrollment (запись на курс), используя 3-4 структуры:
	// например: Student, Course, ContactInfo, CourseEnrollment.

	// Требования:
	// у CourseEnrollment должны быть собственные поля (например, EnrollmentID, Status);
	// хотя бы одна структура должна быть вложенным полем;
	// хотя бы одна часть должна подключаться через embedding;
	// в main создай один объект и выведи короткий отчет о записи.
	// Проверь, что код остается читаемым: имена полей понятны, роли структур не смешаны.

	enrollment := CourseEnrollment{
		EnrollmentID: 1001,
		Status:       "Active",
		Student: Student{
			ID:   1,
			Name: "Alice",
			ContactInfo: ContactInfo{
				Email: "alice@example.com",
				Phone: "+123456789",
			},
		},
		Course: Course{
			ID:    10,
			Title: "Go Programming",
			Hours: 40,
		},
	}

	fmt.Println("Enrollment Report")
	fmt.Println("------------------")
	fmt.Println("Enrollment ID:", enrollment.EnrollmentID)
	fmt.Println("Student:", enrollment.Name)
	fmt.Println("Email:", enrollment.Email)
	fmt.Println("Course:", enrollment.Title)
	fmt.Println("Status:", enrollment.Status)
}
