package main

import (
	"net/http"
	"fmt"
	"time"
	"os"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"io/ioutil"
	"strings"
)

type HoldersResData struct {
	TotalNum int64 `json:"totalNum"`
	AddressHolders []HoldersItem `json:"addressHolders"`
}

type HoldersItem struct {
	Address string `json:"address"`
	Percentage float64 `json:"percentage"`
	Quantity float64 `json:"quantity"`
	Tag string `json:"tag"`
}

var mysqlDb *sql.DB


func createServer() {
	port := os.Getenv("PORT")

	if len(port) <= 0 {
		port = "8020"
	}

	fmt.Println(time.Now().String(), "port: ", port)

	http.HandleFunc("/api/v1/asset/holders/cos-2e4", func(writer http.ResponseWriter, request *http.Request) {
		page := request.URL.Query().Get("page")
		rows := request.URL.Query().Get("rows")
		fmt.Fprintf(writer, "Hello, page: %s, rows: %s", page, rows)
		fmt.Println(time.Now().String(), "Hello, page: %s, rows: %s", page, rows)
	})

	http.ListenAndServe(":" + port, nil)
}

func createTicker() {
	// getDataFromBinance()
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			getDataFromBinance()
		}
	}()
}

func getDataFromBinance() {
	response, error := http.Get("https://explorer.binance.org/api/v1/asset-holders?page=1&rows=2&asset=COS-2E4")
	if error != nil {
		fmt.Println(time.Now().String(), "responseError: ", error)
	}

	defer response.Body.Close()

	body, bodyError := ioutil.ReadAll(response.Body)
	if error != nil {
		fmt.Println(time.Now().String(), "bodyError: ", bodyError)
	}

	// fmt.Println(time.Now().String(), "string body", string(body))

	var decodedBody HoldersResData
	decodedError := json.Unmarshal(body, &decodedBody)

	if decodedError != nil {
		fmt.Println(time.Now().String(), "decodedError: ", decodedError)
	}

	deleteRe, deleteError := mysqlDb.Query("delete from asset_holders")

	if deleteError != nil {
		fmt.Println(time.Now().String(), "deleteError: ", deleteError)
	}

	defer deleteRe.Close()

	sqlStr := buildInsertSql(decodedBody)

	fmt.Println(time.Now().String(), "sqlStr: ", sqlStr)

	insert, insertError := mysqlDb.Query(sqlStr)

	if insertError != nil {
		fmt.Println(time.Now().String(), "insertError: ", insertError)
	}

	defer insert.Close()

	fmt.Println(time.Now().String(), decodedBody.TotalNum, decodedBody.AddressHolders)
}

func buildInsertSql(data HoldersResData) string {
	list := data.AddressHolders
	var sb strings.Builder

	sb.WriteString("INSERT INTO asset_holders (address,percentage,quantity,tag) VALUES")

	for i := 0; i < len(list); i++ {
		item := list[i]

		if i == len(list) - 1 {
			sb.WriteString(fmt.Sprintf("(%s,%f,%f,%s);", item.Address, item.Quantity, item.Percentage, item.Tag))
		} else {
			sb.WriteString(fmt.Sprintf("(%s,%f,%f,%s),", item.Address, item.Quantity, item.Percentage, item.Tag))
		}
	}
	return sb.String()
}

func connectDb() {
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlPort := os.Getenv("MYSQL_PORT")
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlDbName := os.Getenv("MYSQL_DB_NAME")

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlPort, mysqlDbName)

	var dbError error
	mysqlDb, dbError = sql.Open("mysql", dataSourceName)

	if dbError != nil {
		fmt.Println(time.Now().String(), "dbError: ", dbError)
	}

	// defer mysqlDb.Close()
}

func main() {
	connectDb()
	go createTicker()
	createServer()
}

// asset_holders
