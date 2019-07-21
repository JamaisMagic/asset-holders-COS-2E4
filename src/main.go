package main

import (
	"net/http"
	"fmt"
	"time"
	"io/ioutil"
	"os"
)

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
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			getFromBinance()
		}
	}()
}

func getFromBinance() {
	response, error := http.Get("https://explorer.binance.org/api/v1/asset-holders?page=1&rows=2&asset=COS-2E4")
	if error != nil {

	}

	defer response.Body.Close()

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {

	}

	fmt.Println(string(body))
}

func main() {
	go createTicker()
	createServer()
}

// 127.0.0.1:6379 explorer_picoluna_com asset_holders
