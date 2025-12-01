package core

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// DB is the database connection.
var DB *sql.DB

// InitDB initializes the database connection.
func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	fmt.Println("Successfully connected to the database!")

	if err := InitSchema(); err != nil {
		return fmt.Errorf("failed to init schema: %w", err)
	}

	if err := SeedData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	return nil
}

// InitSchema creates the necessary tables if they don't exist.
func InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			category TEXT NOT NULL,
			manufacturer TEXT NOT NULL,
			availability TEXT NOT NULL,
			price INT NOT NULL,
			old_price INT,
			description TEXT NOT NULL,
			features TEXT[], -- Using Postgres array type
			image TEXT NOT NULL,
			stock INT NOT NULL,
			rating FLOAT NOT NULL,
			reviews INT NOT NULL,
			sku TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS cart_items (
			id SERIAL PRIMARY KEY,
			session_id TEXT NOT NULL,
			product_id TEXT NOT NULL REFERENCES products(id),
			quantity INT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(session_id, product_id)
		);`,
		`CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			order_number TEXT UNIQUE NOT NULL,
			customer_type TEXT NOT NULL,
			name TEXT NOT NULL,
			phone TEXT NOT NULL,
			email TEXT NOT NULL,
			company_name TEXT,
			bin TEXT,
			address TEXT NOT NULL,
			total_amount INT NOT NULL,
			status TEXT DEFAULT 'new',
			created_at TIMESTAMP DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id TEXT NOT NULL REFERENCES orders(id),
			product_id TEXT NOT NULL REFERENCES products(id),
			quantity INT NOT NULL,
			price_at_purchase INT NOT NULL
		);`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// SeedData inserts initial data if products table is empty.
func SeedData() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	log.Println("Seeding products...")

	// Mock products data
	products := []struct {
		ID           string
		Name         string
		Category     string
		Manufacturer string
		Availability string
		Price        int
		OldPrice     *int
		Description  string
		Features     []string
		Image        string
		Stock        int
		Rating       float64
		Reviews      int
		SKU          string
	}{
		{
			ID:           "1",
			Name:         "Респиратор 3M 7502",
			Category:     "Респираторы",
			Manufacturer: "3M",
			Availability: "in-stock",
			Price:        12500,
			Description:  "Полумаска 3M™ серии 7500 – это уникальное сочетание плотного прилегания, мягкости и комфорта.",
			Features:     []string{"Материал: Силикон", "Тип: Полумаска", "Защита: Газы и пары"},
			Image:        "https://noble-group.vercel.app/images/product1.jpg",
			Stock:        100,
			Rating:       4.8,
			Reviews:      124,
			SKU:          "3M-7502",
		},
		{
			ID:           "2",
			Name:         "Перчатки Ansell HyFlex",
			Category:     "Перчатки",
			Manufacturer: "Ansell",
			Availability: "in-stock",
			Price:        5000,
			Description:  "Перчатки HyFlex® 11-800 обеспечивают легендарную точность манипуляций и комфорт.",
			Features:     []string{"Материал: Нитрил", "Размер: 9", "Защита: Механическая"},
			Image:        "https://noble-group.vercel.app/images/product2.jpg",
			Stock:        200,
			Rating:       4.5,
			Reviews:      80,
			SKU:          "AN-11-800",
		},
		{
			ID:           "3",
			Name:         "Каска MSA V-Gard",
			Category:     "Каски",
			Manufacturer: "MSA Safety",
			Availability: "pre-order",
			Price:        8000,
			Description:  "Защитная каска V-Gard® из полиэтилена высокой плотности (HDPE) с УФ-ингибитором.",
			Features:     []string{"Материал: HDPE", "Вентиляция: Нет", "Цвет: Белый"},
			Image:        "https://noble-group.vercel.app/images/product3.jpg",
			Stock:        0,
			Rating:       4.7,
			Reviews:      50,
			SKU:          "MSA-VGARD",
		},
		{
			ID:           "4",
			Name:         "Очки Uvex Skyper",
			Category:     "Очки",
			Manufacturer: "Uvex",
			Availability: "in-stock",
			Price:        3500,
			Description:  "Открытые очки с панорамным обзором и отличной защитой.",
			Features:     []string{"Покрытие: supravision", "Цвет линз: Прозрачный", "Защита от УФ: Да"},
			Image:        "https://noble-group.vercel.app/images/product4.jpg",
			Stock:        150,
			Rating:       4.6,
			Reviews:      95,
			SKU:          "UV-9195",
		},
		{
			ID:           "5",
			Name:         "Костюм Tyvek Classic",
			Category:     "Спецодежда",
			Manufacturer: "Tyvek",
			Availability: "in-stock",
			Price:        15000,
			Description:  "Комбинезон Tyvek® Classic Xpert обеспечивает превосходную защиту от твердых частиц.",
			Features:     []string{"Тип: Комбинезон", "Размер: L", "Защита: Химзащита"},
			Image:        "https://noble-group.vercel.app/images/product5.jpg",
			Stock:        50,
			Rating:       4.9,
			Reviews:      200,
			SKU:          "TY-CLASSIC",
		},
	}

	for _, p := range products {
		// Postgres array literal format: {val1,val2}
		// Quick and dirty formatting for string array
		featuresStr := "{"
		for i, f := range p.Features {
			if i > 0 {
				featuresStr += ","
			}
			featuresStr += fmt.Sprintf("\"%s\"", f)
		}
		featuresStr += "}"

		_, err := DB.Exec(`INSERT INTO products (
			id, name, category, manufacturer, availability, price, old_price, description, features, image, stock, rating, reviews, sku
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
			p.ID, p.Name, p.Category, p.Manufacturer, p.Availability, p.Price, p.OldPrice, p.Description, featuresStr, p.Image, p.Stock, p.Rating, p.Reviews, p.SKU)
		if err != nil {
			return err
		}
	}
	return nil
}
