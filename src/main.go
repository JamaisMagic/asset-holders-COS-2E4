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
	"github.com/go-redis/redis"
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

type VisitCount struct {
	Ip string `json:"ip"`
	Count int32 `json:"count"`
	Message string `json:"message"`
}

const visitCountPrefix = "visit:count:"

var mysqlDb *sql.DB
var redisClient *redis.Client


func createServer() {
	port := os.Getenv("PORT")

	if len(port) <= 0 {
		port = "8020"
	}

	http.HandleFunc("/api/v1/asset/holders/cos-2e4/item", handlerCos2e4Item)
	http.HandleFunc("/api/v1/common/visit-count", handleVisitCount)

	fmt.Println(time.Now().String(), "Listening port: ", port)
	http.ListenAndServe(":" + port, nil)
}

func createTicker() {
	ticker := time.NewTicker(1800 * time.Second)
	go func() {
		for range ticker.C {
			getDataFromBinance()
		}
	}()
}

func getDataFromBinance() {
	request, requestError := http.NewRequest("GET", "https://explorer.binance.org/api/v1/asset-holders?page=1&rows=0&asset=COS-2E4", nil)

	if requestError != nil {
		panic(requestError)
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	request.Header.Set("Referer", "https://explorer.binance.org/asset/holders/COS-2E4")

	client := &http.Client{}

	response, responseError := client.Do(request)

	if responseError != nil {
		panic(responseError)
	}

	defer response.Body.Close()

	body, bodyError := ioutil.ReadAll(response.Body)

	if bodyError != nil {
		panic(bodyError)
	}

	var decodedBody HoldersResData
	decodedError := json.Unmarshal(body, &decodedBody)

	if decodedError != nil {
		panic(decodedError)
	}

	updateNewData(decodedBody)
}

func queryItemByAddress(address string) (HoldersItem, error) {
	var itemRow HoldersItem
	queryRow := mysqlDb.QueryRow("select address, quantity, percentage, tag from asset_holders where address = ?", address)
	err := queryRow.Scan(&itemRow.Address, &itemRow.Quantity, &itemRow.Percentage, &itemRow.Tag)

	if err != nil {
		return itemRow, err
	}

	return itemRow, nil
}

func updateNewData(decodedBody HoldersResData) {
	tx, txError := mysqlDb.Begin()

	if txError != nil {
		panic(txError)
	}

	_, deleteError := tx.Exec("delete from asset_holders;")

	if deleteError != nil {
		tx.Rollback()
		panic(deleteError)
	}

	sqlStr := buildInsertSql(decodedBody)

	_, insertError := tx.Exec(sqlStr)

	if insertError != nil {
		tx.Rollback()
		panic(insertError)
	}

	commitError := tx.Commit()

	if commitError != nil {
		tx.Rollback()
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
		panic(dbError)
	}
}

func connectRedis() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	redisClient = redis.NewClient(&redis.Options {
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "",
		DB: 0,
	})

	pong, err = redisClient.Ping().Result()

	if err != nil {
		panic(err)
	}

	fmt.Println(time.Now().String(), "Redis connected: ", pong)
}

func handlerCos2e4Item(writer http.ResponseWriter, request *http.Request) {
	address := request.URL.Query().Get("address")

	if len(address) <= 0 {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(nil)
		return
	}

	item, itemError := queryItemByAddress(address)

	if itemError != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(nil)
		return
	}

	resJson, err := json.Marshal(item)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(resJson)
}

func handleVisitCount(writer http.ResponseWriter, request *http.Request) {
	remoteIp := request.RemoteAddr
	xForwardedFor := strings.Split(request.Header.Get("X-Forwarded-For"), ",")[0]
	ip := xForwardedFor

	if len(ip) <= 0 {
		ip = remoteIp
	}

	key := visitCountPrefix + ip

	err := redisClient.incr(key).Err()

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	count, err := redisClient.Get(key).Result()

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	var resData VisitCount

	resData.Count = count
	resData.Ip = ip
	resData.Message = fmt.Sprintf("Your ip address is %s, you'v visit %s times", ip, count)

	resJson, err := json.Marshal(resData)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(resJson)
}

func main() {
	connectDb()
	connectRedis()
	go createTicker()

	defer func() {
		if mysqlDb != nil {
			mysqlDb.Close()
		}
	}()

	createServer()
}
