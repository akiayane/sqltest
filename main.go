package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

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
}

type Testpack struct {
	users    User
	products Product
}

type Config struct {
	Dsn          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type application struct {
	db       *sql.DB
	testpack interface{}
}

func NewSQLTest(config string, testpack interface{}) application {

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
		db:       db,
		testpack: testpack,
	}

	return application
}

func (app application) Iterate() {
	tablenames := []string{}

	val := reflect.ValueOf(app.testpack).Elem()
	for i := 0; i < val.NumField(); i++ {
		tablenames = append(tablenames, val.Type().Field(i).Name)
	}
	testpack_reflection := reflect.ValueOf(app.testpack)

	for i, v := range tablenames {
		current_tablename := v
		var fieldnames []string
		var data []interface{}

		//getting required variables
		nest := reflect.Indirect(testpack_reflection).FieldByIndex([]int{i})

		for i := 0; i < nest.NumField(); i++ {
			fieldnames = append(fieldnames, nest.Type().Field(i).Name)
		}

		for _, v := range fieldnames {
			current_field := nest.FieldByName(v)
			ty := current_field.Type().String()
			if ty == "int" {
				current_field := nest.FieldByName(v)
				data = append(data, int(current_field.Int()))
			} else if ty == "string" {
				current_field := nest.FieldByName(v)
				data = append(data, current_field.String())
			}
		}

		fmt.Println(current_tablename)
		/*fmt.Println(fieldnames)
		fmt.Println(data)*/
		//-----------------------------------------------------------------------------

		var fields string
		var values string

		for _, s := range fieldnames {
			fields += s + ","
		}
		for _, s := range data {
			str := fmt.Sprintf("%v", s)
			values += "'" + str + "',"
		}

		fields = strings.ToLower(strings.TrimSuffix(fields, ","))
		values = strings.ToLower(strings.TrimSuffix(values, ","))

		fmt.Println(fields)
		fmt.Println(values)

		stmt := "INSERT INTO " + current_tablename + " (" + fields + ") VALUES(" + values + ");"
		_, err := app.db.Exec(stmt)
		if err != nil {
			log.Printf("Unable to INSERT: %v\n", err)
		}

		var where string

		for i, s := range fieldnames {
			str := fmt.Sprintf("%v", data[i])
			where += s + "='" + str + "' AND "
		}

		where = strings.ToLower(strings.TrimRight(where, " AND "))

		stmt = "SELECT " + fields + " from " + current_tablename + " WHERE " + where + ";"
		rows, err := app.db.Query(stmt)
		if err != nil {
			log.Printf("Unable to SELECT: %v\n", err)
		}

		defer rows.Close()

		if !rows.Next() {
			fmt.Println("no such data in table " + current_tablename)
		}

		stmt = "DELETE FROM " + current_tablename + " WHERE + " + where + ";"
		ct, err := app.db.Exec(stmt)
		if err != nil {
			log.Printf("Unable to DELETE: %v\n", err)
		}

		if temp, _ := ct.RowsAffected(); temp == 0 {
			// Work with Error
			log.Printf("no affected rows")
		}
	}
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

	application := NewSQLTest("./config.json", testpack)
	application.Iterate()
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
