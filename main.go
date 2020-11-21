package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "database/sql"
    "log"
    "os"
    _ "github.com/lib/pq"
)

type Customer struct {
    ID int `json:"id"`
    Name string `json:"name"`
    Email string `json:"email"`
    Status string `json:"status"`
}

var DB *sql.DB

type Conn interface {
    conn() *sql.DB
    initTable(*sql.DB)
}

func conn() *sql.DB {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal("DB connection error", err)
    }
    return db
}

func initTable(db *sql.DB) {
    createDb := `CREATE TABLE IF NOT EXISTS customers(
        id SERIAL PRIMARY KEY,
        name TEXT,
        email TEXT,
        status TEXT
    );`

    _, err := db.Exec(createDb)
    if err != nil {
        log.Fatal("Can't create table", err)
    }
}

func init() {
    DB = conn()
    initTable(DB)
}

func getCustomersHandler(c *gin.Context) {
    statement := "SELECT id, name, email, status FROM customers"

    stmt, err := DB.Prepare(statement)
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
        return
    }
    rows, err := stmt.Query()
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
        return
    }

    custs := []Customer{}
    for rows.Next() {
        cust := Customer{}
        rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
        custs = append(custs, cust)
    }
    c.JSON(http.StatusOK, custs)
}

func getCustomersByIDHandler(c *gin.Context) {
    statement := "SELECT id, name, email, status FROM customers WHERE id = " + c.Param("id")
    stmt, err := DB.Prepare(statement)
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
        return
    }
    rows, err := stmt.Query()
    if err != nil {
        c.JSON(http.StatusInternalServerError, err)
        return
    }

    var cust Customer
    for rows.Next() {
        rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
    }
    if cust.ID == 0 {
        c.JSON(http.StatusOK, gin.H{})
    } else {
        c.JSON(http.StatusOK, cust)
    }
}

func createCustomersHandler(c *gin.Context) {
    var cust Customer
    if err := c.ShouldBindJSON(&cust); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{ "error": err.Error() })
        return
    }

    row := DB.QueryRow("INSERT INTO customers (name, email, status) VALUES ($1, $2, $3) RETURNING id", cust.Name, cust.Email, cust.Status)
    err := row.Scan(&cust.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{ "error": err.Error() })
        return
    }
    c.JSON(http.StatusCreated, cust)
}

func updateCustomersHandler(c *gin.Context) {
    var cust Customer
    if err := c.ShouldBindJSON(&cust); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{ "error": err.Error() })
        return
    }

    row := DB.QueryRow("UPDATE customers SET name = $2, email = $3, status = $4 WHERE id = $1 RETURNING id", c.Param("id"), cust.Name, cust.Email, cust.Status)

    err := row.Scan(&cust.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{ "error": err.Error() })
        return
    }
    c.JSON(http.StatusOK, cust)
}

func deleteCustomersHandler(c *gin.Context) {
    DB.QueryRow("DELETE FROM customers WHERE id = $1", c.Param("id"))
    c.JSON(http.StatusOK, gin.H{ "message": "customer deleted" })
}

func main() {
    defer DB.Close()

    srv := gin.Default()
    srv.GET("/customers", getCustomersHandler)
    srv.GET("/customers/:id", getCustomersByIDHandler)
    srv.POST("/customers", createCustomersHandler)
    srv.PUT("/customers/:id", updateCustomersHandler)
    srv.DELETE("/customers/:id", deleteCustomersHandler)
    srv.Run(os.Getenv("PORT"))
}
