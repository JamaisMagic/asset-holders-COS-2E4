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
	"strconv"
	_ "net/http/pprof"
	"log"
	"runtime/pprof"
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
	http.HandleFunc("/api/v1/test/cpu", handleCpuTest)

	log.Println(http.ListenAndServe(":" + port, nil))
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

	fmt.Println(time.Now().String(), "MySql connected.")
}

func connectRedis() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	redisClient = redis.NewClient(&redis.Options {
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "",
		DB: 0,
	})

	pong, err := redisClient.Ping().Result()

	if err != nil {
		panic(err)
	}

	fmt.Println(time.Now().String(), "Redis connected: ", pong)
}

func handlerCos2e4Item(writer http.ResponseWriter, request *http.Request) {
	address := request.URL.Query().Get("address")

	if len(address) <= 0 {
		fmt.Println(time.Now().String(), "No address: ", address)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(nil)
		return
	}

	item, itemError := queryItemByAddress(address)

	if itemError != nil {
		fmt.Println(time.Now().String(), itemError)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(nil)
		return
	}

	resJson, err := json.Marshal(item)

	if err != nil {
		fmt.Println(time.Now().String(), err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(resJson)
}

func handleVisitCount(writer http.ResponseWriter, request *http.Request) {
	ipAddress := request.RemoteAddr
	xForwardedFor := request.Header.Get("X-Forwarded-For")
	xForwardedForIps := strings.Split(xForwardedFor, ",")

	if len(xForwardedForIps) >= 0 {
		ipAddress = xForwardedForIps[0]
	}

	key := visitCountPrefix + ipAddress
	err := redisClient.Incr(key).Err()

	if err != nil {
		fmt.Println(time.Now().String(), err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	redisClient.Expire(key, 60 * time.Second).Err()
	count, err := redisClient.Get(key).Result()

	if err != nil {
		fmt.Println(time.Now().String(), err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	var resData VisitCount
	count64, count64Err := strconv.ParseInt(count, 10, 32)

	if count64Err != nil {
		fmt.Println(time.Now().String(), count64Err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	resData.Count = int32(count64)
	resData.Ip = ipAddress
	resData.Message = fmt.Sprintf("Your ip address is %s, you'v requested this url %s times. The counter will be reset in 1 minute if you do nothing.", ipAddress, count)
	resJson, err := json.Marshal(resData)

	if err != nil {
		fmt.Println(time.Now().String(), err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(resJson)
}

func handleCpuTest(writer http.ResponseWriter, request *http.Request) {
	pprof.StartCPUProfile(os.Stdout)
	defer pprof.StopCPUProfile()

	fmt.Println(time.Now().String(), "CPU test started.")
	count := createHeavy(10000 * 10000)
	fmt.Println(time.Now().String(), "CPU test ended.")

	resJson, err := json.Marshal(count)

	if err != nil {
		fmt.Println(time.Now().String(), err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(nil)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(resJson)
}

func createHeavy(size int) int {
	if size <= 0 {
		size = 1000 * 1000
	}

	count := 0

	for i := 0; i < size; i++ {
		count += i
	}

	return count
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
