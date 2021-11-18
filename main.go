package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id int
	Name string
	Surname string
	Password string
	Email string
}

type Product struct{
	Id int
	Name string
	Count int
	Description string
}

type Testpack struct {
	users User
	products Product
}

type Config struct {
	Dsn string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime string
}

type application struct {
	db *sql.DB
	testpack interface{}
}

func NewSQLTest(config string, testpack interface{}) application{

	conf, err := os.Open(config)
	if err != nil {
		fmt.Println(err)
	}
	defer conf.Close()

	var cfg Config
	byteValue, _ := ioutil.ReadAll(conf)
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		fmt.Println(err)
	}

	db, err := openDB(cfg)
	if err != nil {
		fmt.Println(err)
	}

	application := application{
		db: db,
		testpack: testpack,
	}

	return application
}

func(app application) Iterate(){
	tablenames := []string{}

	val := reflect.ValueOf(app.testpack).Elem()
	for i:=0; i<val.NumField();i++{
		tablenames = append(tablenames, val.Type().Field(i).Name)
	}
	testpack_reflection := reflect.ValueOf(app.testpack)

	for i, v := range tablenames{
		current_tablename := v
		var fieldnames []string
		var data []interface{}

		//getting required variables
		nest := reflect.Indirect(testpack_reflection).FieldByIndex([]int{i})

		for i:=0; i<nest.NumField();i++{
			fieldnames = append(fieldnames, nest.Type().Field(i).Name)
		}

		for _, v := range fieldnames{
			current_field := nest.FieldByName(v)
			ty := current_field.Type().String()
			fmt.Println(ty)
			if ty =="int"{
				current_field := nest.FieldByName(v)
				data = append(data, int(current_field.Int()))
			} else if ty=="string"{
				current_field := nest.FieldByName(v)
				data = append(data, current_field.String())
			}
		}

		fmt.Println(current_tablename)
		fmt.Println(fieldnames)
		fmt.Println(data)
		//-----------------------------------------------------------------------------
	}
}

func main() {

	testpack := &Testpack{
		users:    User{
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

	application := NewSQLTest("./config.json", testpack)
	application.Iterate()
	/*
	conf, err := os.Open("./config.json")
	if err != nil {
		fmt.Println(err)
	}
	defer conf.Close()

	var config Config
	byteValue, _ := ioutil.ReadAll(conf)
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	user := &User{
		Id:       2,
		Name:     "Gallardo",
		Surname:  "Migelyan",
		Password: "alibeksmom",
		Email:    "gala@gmail.com",
	}
	val := reflect.ValueOf(user).Elem()
	fields := []string{}
	for i:=0; i<val.NumField();i++{
		fields = append(fields, val.Type().Field(i).Name)
	}
	fmt.Println(fields)
	fmt.Println(user)*/
}


func openDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.Dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

