package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

const TIMEOUT = 2 * time.Minute
const TIMEOUT_STR = "2m"

func main() {
	dbName := os.Getenv("MYSQL_DB")
	mysqlConfig := createMySQLConfig(
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		dbName,
		"1024",
		map[string]string{"tls": "custom"},
	)

	rootCertPool := x509.NewCertPool()
	pem, err := os.ReadFile("public.crt")
	if err != nil {
		log.Fatalf("Error loading CA cert: %v", err)
	}
	rootCertPool.AppendCertsFromPEM(pem)

	err = mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs:            rootCertPool,
		InsecureSkipVerify: false,
	})
	if err != nil {
		log.Fatalf("Error registering TLS Config: %v", err)
	}

	sqlConfig := mysqlConfig.FormatDSN()
	db, err := sql.Open("mysql", sqlConfig)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}

	fmt.Println("Database created successfully!")
}

func createMySQLConfig(user, password, mysqlServiceHost, mysqlServicePort,
	dbName, mysqlGroupConcatMaxLen string, mysqlExtraParams map[string]string,
) *mysql.Config {
	params := map[string]string{
		"charset":              "utf8",
		"parseTime":            "True",
		"loc":                  "Local",
		"group_concat_max_len": mysqlGroupConcatMaxLen,
		"timeout":              TIMEOUT_STR,
		"readTimeout":          TIMEOUT_STR,
		"writeTimeout":         TIMEOUT_STR,
	}

	for k, v := range mysqlExtraParams {
		params[k] = v
	}

	return &mysql.Config{
		User:                 user,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", mysqlServiceHost, mysqlServicePort),
		Params:               params,
		DBName:               dbName,
		AllowNativePasswords: true,
	}
}
