package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

const employeePath = "employee"
const apiBasePath = "/api"

type Employee struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last"`
	Age       int    `json:"age"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Cid       string `json:"cid"`
	Position  string `json:"position"`
}

func SetupDB() {
	var err error
	Db, err = sql.Open("mysql", "root@tcp(127.0.0.1:3306)/coursedb")

	if err != nil {
		log.Fatal(err)
	}
	Db.SetConnMaxLifetime(time.Minute * 3)
	Db.SetMaxOpenConns(10)
	Db.SetMaxIdleConns(10)
	fmt.Println("Connect database successfull")
}

// getEmployeeList get list of employee from table
func getEmployeeList() ([]Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	results, err := Db.QueryContext(ctx, `SELECT * FROM employee`)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer results.Close()
	employees := make([]Employee, 0)
	for results.Next() {
		var employee Employee
		results.Scan(
			&employee.ID,
			&employee.FirstName,
			&employee.LastName,
			&employee.Age,
			&employee.Email,
			&employee.Phone,
			&employee.Cid,
			&employee.Position,
		)
		employees = append(employees, employee)
		fmt.Println("Get employee list successfull")
	}
	return employees, nil
}

// insertEmployee insert one employee to table
func insertEmployee(employee Employee) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := Db.ExecContext(ctx, `INSERT INTO employee (id, first_name, last_name, age, email, phone, cid, position) VALUE (?, ?, ?, ?, ?, ?, ?, ?)`,
		employee.ID, employee.FirstName, employee.LastName, employee.Age, employee.Email, employee.Phone, employee.Cid, employee.Position)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	insertID, err := result.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	fmt.Println("Insert employee successfull")
	return int(insertID), nil
}

// getEmployee get one employee from table
func getEmployee(id int) (*Employee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := Db.QueryRowContext(ctx, `SELECT * FROM employee WHERE id = ?`,
		id)
	employee := &Employee{}
	err := row.Scan(
		&employee.ID,
		&employee.FirstName,
		&employee.LastName,
		&employee.Age,
		&employee.Email,
		&employee.Phone,
		&employee.Cid,
		&employee.Position,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.Println(err)
		return nil, err
	}
	fmt.Println("Get one employee successfull")
	return employee, nil
}

// updateEmployee update one employee with id
func updateEmployee(id int, employee Employee) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// UPDATE Customers
	// SET ContactName = 'Alfred Schmidt', City= 'Frankfurt'
	// WHERE CustomerID = 1;
	result, err := Db.ExecContext(ctx, `UPDATE employee SET id = ?, first_name = ?, last_name = ?, age = ?, email = ?, phone = ?, cid = ?, position = ? WHERE id = ?`,
		employee.ID, employee.FirstName, employee.LastName, employee.Age, employee.Email, employee.Phone, employee.Cid, employee.Position, id)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	insertID, err := result.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	fmt.Println("Update employee successfull")
	return int(insertID), nil
}

// removeEmployee remove one employee with id
func removeEmployee(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Db.ExecContext(ctx, `DELETE FROM employee WHERE id = ?`, id)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	fmt.Println("Remove employee successfull")
	return nil
}

func handleEmployee(w http.ResponseWriter, r *http.Request) {
	urlPathSegments := strings.Split(r.URL.Path, fmt.Sprintf("%s/", employeePath))
	if len(urlPathSegments[1:]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ID, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1])
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		employee, err := getEmployee(ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if employee == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		j, err := json.Marshal(employee)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = w.Write(j)
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodPut:
		var employee Employee
		err := json.NewDecoder(r.Body).Decode(&employee)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = updateEmployee(ID, employee)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case http.MethodDelete:
		err := removeEmployee(ID)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func handleEmployees(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		employeeList, err := getEmployeeList()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		j, err := json.Marshal(employeeList)
		if err != nil {
			log.Fatal(err)
		}
		_, err = w.Write(j)
		if err != nil {
			log.Fatal(err)
		}
	case http.MethodPost:
		var employee Employee
		err := json.NewDecoder(r.Body).Decode(&employee)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ID, err := insertEmployee(employee)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"id}: %d}`, ID)))
	case http.MethodOptions:
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Lengh")
		handler.ServeHTTP(w, r)
	})
}

func setupRoutes(apiBasePath string) {
	employeeHandler := http.HandlerFunc(handleEmployee)
	http.Handle(fmt.Sprintf("%s/%s/", apiBasePath, employeePath), corsMiddleware(employeeHandler))
	employeesHandler := http.HandlerFunc(handleEmployees)
	http.Handle(fmt.Sprintf("%s/%s", apiBasePath, employeePath), corsMiddleware(employeesHandler))
}

func main() {
	SetupDB()
	setupRoutes(apiBasePath)
	log.Fatal(http.ListenAndServe(":5000", nil))
}