package sqltest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

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

func (app application) MainSQLTest() {
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

		//fmt.Println(current_tablename)
		/*fmt.Println(fieldnames)
		fmt.Println(data)*/
		//-----------------------------------------------------------------------------

		errs := make(map[string]error)

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

		//fmt.Println(fields)
		//fmt.Println(values)

		stmt := "INSERT INTO " + current_tablename + " (" + fields + ") VALUES(" + values + ");"
		_, err := app.db.Exec(stmt)
		if err != nil {
			log.Printf("Unable to INSERT: %v\n", err)
			errs["Unable to INSERT: "] = err
		} else {
			log.Printf("Successful INSERT to table %v\n", current_tablename)
			log.Printf("Query: %v\n", stmt)

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
			errs["Unable to SELECT: "] = err
		} else {
			log.Printf("Successful SELECT from table %v\n", current_tablename)
			log.Printf("Query: %v\n", stmt)
		}

		defer func(rows *sql.Rows) {
			if rows != nil {
				err := rows.Close()
				if err != nil {

				}
			}
		}(rows)

		if rows != nil {
			if !rows.Next() {
				fmt.Println("no such data in table " + current_tablename)
			}
		}

		stmt = "DELETE FROM " + current_tablename + " WHERE " + where + ";"
		ct, err := app.db.Exec(stmt)
		if err != nil {
			log.Printf("Unable to DELETE: %v\n", err)
			errs["Unable to DELETE: "] = err
		} else {
			log.Printf("Successful DELETE from table %v\n", current_tablename)
			log.Printf("Query: %v\n", stmt)
		}

		if ct != nil {
			if temp, _ := ct.RowsAffected(); temp == 0 {
				// Work with Error
				log.Printf("no affected rows")
				err = errors.New("no affected rows")
				errs["Unable to DELETE: "] = err
			}
		}

		if len(errs) == 0 {
			log.Printf("No errors in table: %v\n", current_tablename)
		} else {
			log.Printf("Errors in table: %v\n", current_tablename)
			for _, e := range errs {
				log.Printf("%v\n", e)
			}
		}
	}
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
