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
	Quantity float64 `json:"quantity"`
	Percentage float64 `json:"percentage"`
	Tag string `json:"tag"`
}

var mysqlDb *sql.DB


func createServer() {
	port := os.Getenv("PORT")

	if len(port) <= 0 {
		port = "8020"
	}

	http.HandleFunc("/api/v1/asset/holders/cos-2e4", func(writer http.ResponseWriter, request *http.Request) {
		page := request.URL.Query().Get("page")
		rows := request.URL.Query().Get("rows")


		fmt.Fprintf(writer, "Hello, page: %s, rows: %s", page, rows)
	})

	http.HandleFunc("/api/v1/asset/holders/cos-2e4/item", func(writer http.ResponseWriter, request *http.Request) {
		address := request.URL.Query().Get("address")
		item := queryItemByAddress(address)
		resJson, err := json.Marshal(item)

		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		writer.Write(resJson)
	})

	fmt.Println(time.Now().String(), "Listening port: ", port)

	http.ListenAndServe(":" + port, nil)
}

func createTicker() {
	getDataFromBinance()

	ticker := time.NewTicker(1800 * time.Second)
	go func() {
		for range ticker.C {
			getDataFromBinance()
		}
	}()
}

func getDataFromBinance() {
	response, responseError := http.Get("https://explorer.binance.org/api/v1/asset-holders?page=1&rows=0&asset=COS-2E4")

	if responseError != nil {
		fmt.Println(time.Now().String(), "responseError: ", responseError)
		panic(responseError)
	}

	defer response.Body.Close()

	body, bodyError := ioutil.ReadAll(response.Body)

	if bodyError != nil {
		fmt.Println(time.Now().String(), "bodyError: ", bodyError)
		panic(bodyError)
	}

	var decodedBody HoldersResData
	decodedError := json.Unmarshal(body, &decodedBody)

	if decodedError != nil {
		fmt.Println(time.Now().String(), "decodedError: ", decodedError)
		panic(decodedError)
	}

	updateNewData(decodedBody)
}

func queryItemByAddress(address string) HoldersItem {
	var itemRow HoldersItem
	queryRow := mysqlDb.QueryRow("select address, quantity, percentage, tag from asset_holders where address = ?", address)
	err := queryRow.Scan(&itemRow.Address, &itemRow.Quantity, &itemRow.Percentage, &itemRow.Tag)

	if err != nil {
		panic(err)
	}

	return itemRow
}

func updateNewData(decodedBody HoldersResData) {
	tx, txError := mysqlDb.Begin()

	if txError != nil {
		fmt.Println(time.Now().String(), "txError: ", txError)
		panic(txError)
	}

	_, deleteError := tx.Exec("delete from asset_holders;")

	if deleteError != nil {
		tx.Rollback()
		fmt.Println(time.Now().String(), "deleteError: ", deleteError)
		panic(deleteError)
	}

	sqlStr := buildInsertSql(decodedBody)

	_, insertError := tx.Exec(sqlStr)

	if insertError != nil {
		tx.Rollback()
		fmt.Println(time.Now().String(), "insertError: ", insertError)
		panic(insertError)
	}

	commitError := tx.Commit()

	if commitError != nil {
		tx.Rollback()
		fmt.Println(time.Now().String(), "commitError: ", commitError)
		panic(commitError)
	}
}

func buildInsertSql(data HoldersResData) string {
	list := data.AddressHolders
	listLen := len(list)
	var sb strings.Builder

	sb.WriteString("INSERT INTO asset_holders (address,quantity,percentage,tag) VALUES ")

	for i := 0; i < listLen; i++ {
		item := list[i]

		// should encode values or using prepare statement for security reason.
		if i == listLen - 1 {
			sb.WriteString(fmt.Sprintf(`("%s","%.8f","%.4f","%s");`, item.Address, item.Quantity, item.Percentage, item.Tag))
		} else {
			sb.WriteString(fmt.Sprintf(`("%s","%.8f","%.4f","%s"),`, item.Address, item.Quantity, item.Percentage, item.Tag))
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
		panic(dbError)
	}
}

func main() {
	connectDb()
	go createTicker()

	defer func() {
		if mysqlDb != nil {
			mysqlDb.Close()
		}
	}()

	createServer()
}
