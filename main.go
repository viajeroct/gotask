package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math/rand"
)

// табличка студентов
type Student struct {
	Id       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Nickname string `json:"nickname" db:"nickname"`
	Money    int    `json:"money" db:"money"`
}

// табличка школ и какой-то информации о нем
type School struct {
	Id       int    `json:"id" db:"id"`
	Info     string `json:"info" db:"info"`
	Nickname string `json:"nickname" db:"nickname"`
	School   int    `json:"school" db:"school"`
	Money    int    `json:"money" db:"money"`
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// просто какой-то генератор строчек
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// это пример работы транзакции
// вызов и пояснение ниже в main
func setValueToTables(value int, db *sqlx.DB, nickname string, shouldPanic bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
		}
	}()

	tx, err := db.Begin()
	if err != nil {
		return
	}
	tx.Exec("update students set money=? where nickname=?;", value, nickname)
	if shouldPanic {
		panic("paaanic")
	}
	tx.Exec("update info set money=? where nickname=?;", value, nickname)
	err = tx.Commit()
	if err != nil {
		return
	}
}

/*
*
наверное не очень красиво
но постарался потыкать что было в лекции
^~^
*/
func main() {
	db, err := sqlx.Open("sqlite3", "./db/university.db")

	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	if err != nil {
		log.Fatal(err)
	}

	// удаляю чтобы два запуска подряд друг на друга не влияли
	_, err = db.Exec("DROP TABLE IF EXISTS students;")
	if err != nil {
		log.Fatal(err)
		return
	}

	_, err = db.Exec("DROP TABLE IF EXISTS info;")
	if err != nil {
		log.Fatal(err)
		return
	}

	// создаю две таблички чтобы не руками а в коде
	schema := `
	CREATE TABLE IF NOT EXISTS students (
    	id integer primary key,
		name text not null,
		nickname text not null unique,
		money integer default 0
    );

	CREATE TABLE IF NOT EXISTS info (
    	id integer primary key,
		info text not null,
		nickname text not null unique,
		school integer not null,
		money integer default 0
    );`

	_, err = db.Exec(schema)

	if err != nil {
		log.Fatal(err)
		return
	}

	// просто генерирую никнейм уникальный
	un := make([]string, 3)
	for i := 0; i < 3; i++ {
		un[i] = RandStringBytes(10)
	}

	// добавим информацию в таблички
	for i, name := range []string{"nikita", "veronika", "polina"} {
		_, err = db.Exec("INSERT INTO students (name, nickname) values (?, ?);", name, un[i])
		if err != nil {
			return
		}
	}

	for i, school := range []int{239, 30, 239} {
		_, err = db.Exec("INSERT INTO info (info, nickname, school) values (?, ?, ?);", RandStringBytes(2), un[i], school)
	}

	// тут типа проверяем что без паники посередине оба запроса прошли
	// и деньги мы поставили в обе таблицы
	setValueToTables(100, db, un[0], false)

	// а тут случилась паника и мы не смогли сделать commit - все плохо
	// и да - rollback делать не надо
	setValueToTables(300, db, un[0], true)

	var students []Student

	if err = db.Select(&students, "select * from students;"); err != nil {
		log.Fatal(err)
	}

	fmt.Println(students)

	var schools []School

	if err = db.Select(&schools, "select * from info;"); err != nil {
		log.Fatal(err)
	}

	fmt.Println(schools)

	// как ты и просил - сложный запрос с join
	complexQuery := `
	SELECT s.name, i.school, i.info
	FROM students s
	JOIN info i ON s.nickname = i.nickname
	ORDER BY i.school DESC;
    `

	rows, err := db.Query(complexQuery)
	if err != nil {
		log.Fatal(err)
		return
	}

	for rows.Next() {
		var nickname string
		var school int
		var info string

		errs := rows.Scan(&nickname, &school, &info)
		if errs != nil {
			log.Fatal(errs)
			return
		}

		fmt.Println(nickname, school, info)
	}
}

/**
Вывод кода:
paaanic
[{1 nikita dRHwMGMWTm 100} {2 veronika BjOJhUAYaP 0} {3 polina rDYeHHaNBN 0}]
[{1 kj dRHwMGMWTm 239 100} {2 NG BjOJhUAYaP 30 0} {3 nG rDYeHHaNBN 239 0}]
nikita 239 kj
polina 239 nG
veronika 30 NG
*/
