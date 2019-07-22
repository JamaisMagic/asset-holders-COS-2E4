package main

import (
	"net/http"
	"fmt"
	"time"
	"io/ioutil"
	"os"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
)

type HoldersResData struct {
	totalNum int64
	addressHolders []HoldersItem
}

type HoldersItem struct {
	address string
	percentage float64
	quantity float64
	tag string
}


func createServer() {
	port := os.Getenv("PORT")

	if len(port) <= 0 {
		port = "8020"
	}

	fmt.Println("port: ", port)

	http.HandleFunc("/api/v1/asset/holders/cos-2e4", func(writer http.ResponseWriter, request *http.Request) {
		page := request.URL.Query().Get("page")
		rows := request.URL.Query().Get("rows")
		fmt.Fprintf(writer, "Hello, page: %s, rows: %s", page, rows)
		fmt.Println("Hello, page: %s, rows: %s", page, rows)
	})

	http.ListenAndServe(":" + port, nil)
}

func createTicker() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			getDataFromBinance()
		}
	}()
}

func getDataFromBinance() {
	response, error := http.Get("https://explorer.binance.org/api/v1/asset-holders?page=1&rows=2&asset=COS-2E4")
	if error != nil {

	}

	defer response.Body.Close()

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {

	}

	var decodedBody HoldersResData
	decodedError := json.Unmarshal(body, &decodedBody)

	if decodedError != nil {
		fmt.Println("decodedError", decodedError)
	}


	fmt.Println(decodedBody.totalNum, decodedBody.addressHolders)
}

func main() {
	// db, dbError := sql.Open("mysql", "root:123456@/explorer_picoluna_com")
	//
	// if dbError != nil {
	//
	// }
	go createTicker()
	createServer()
}

// asset_holders
