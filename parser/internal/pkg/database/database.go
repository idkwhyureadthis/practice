package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/idkwhyureadthis/practice/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type DB struct {
	connection *sql.DB
}

func SetupDatabase() DB {
	conn, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}
	SetupMigrations(conn)
	return DB{
		connection: conn,
	}
}

func SetupMigrations(conn *sql.DB) {
	err := goose.Up(conn, "internal/migrations")
	if err != nil {
		log.Fatal(err)
	}
}

func (db *DB) SaveToDB(data models.PageData) {
	for _, elem := range data.Jobs {
		var experienceNumber int
		switch experienceString := elem.Experience.Id; experienceString {
		case "noExperience":
			experienceNumber = 0
		case "between1And3":
			experienceNumber = 1
		case "between3And6":
			experienceNumber = 2
		case "moreThan6":
			experienceNumber = 3
		}

		query := fmt.Sprintf(`INSERT INTO vacancies (id, name, url, salary_from, salary_to, currency, experience, employer_name, city_name)
		VALUES (LOWER('%s'), LOWER('%s'), LOWER('%s'), %d, %d, LOWER('%s'), %d, LOWER('%s'), LOWER('%s'))
		`, elem.Id, elem.Name, elem.URL, elem.Salary.SalaryFrom, elem.Salary.SalaryTo, elem.Salary.Currency, experienceNumber, elem.Employer.Name, elem.Area.Name)

		db.connection.Exec(query)
	}
}

func (db *DB) GetFromDB(name, city string, salaryFrom, experience int) models.PageData {
	name = strings.ToLower(name)
	city = strings.ToLower(city)

	query := fmt.Sprint(`SELECT * FROM VACANCIES WHERE name LIKE '%`, name, `%' AND city_name LIKE '%`, city, `%' AND (salary_from >= `, salaryFrom, ` OR salary_to >= `, salaryFrom, `) AND experience <= `, experience)
	log.Println(query)
	rows, err := db.connection.Query(query)
	if err != nil {
		log.Println(err)
	}
	resp := models.PageData{}
	for rows.Next() {
		var job models.Job
		err = rows.Scan(&job.Id, &job.Name, &job.URL, &job.Salary.SalaryFrom, &job.Salary.SalaryTo, &job.Salary.Currency, &job.ExperienceInt, &job.Employer.Name, &job.Area.Name)
		if err != nil {
			log.Println(err)
			continue
		}
		resp.Jobs = append(resp.Jobs, job)
	}
	return resp
}
