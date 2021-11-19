package main

import (
	"awesomeProject2/sqltest"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id       int
	Name     string
	Surname  string
	Password string
	Email    string
}

type Product struct {
	Id          int
	Name        string
	Count       int
	Description string
	Price       int
}

type Testpack struct {
	users    User
	products Product
}

func main() {
	testpack := &Testpack{
		users: User{
			Id:       2,
			Name:     "Gallardo",
			Surname:  "Migelyan",
			Password: "alibeksmom",
			Email:    "gala@gmail.com",
		},
		products: Product{
			Id:          13,
			Name:        "T-Shirt",
			Count:       7,
			Description: "Blue T-Shirt",
		},
	}

	application := sqltest.NewSQLTest("./config.json", testpack)
	application.MainSQLTest()
}
