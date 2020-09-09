package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	error            = "error"
	password         = ""
	DB               = 0
	connectionString = "user=postgres dbname=testdb sslmode=disable"
	serverAddress    = "localhost:8081"
	error600         = "Invalid key"
	error601         = "Invalid value"
	error602         = "Invalid text"
	error603         = "Invalid age"
	error604         = "Invalid name"
	error605         = "Table creation failed"
	error606         = "Database connection failed"
	error607         = "Convertion failed"
	error608         = "Increment failed"
)

func getRedisAddress() string {
	return os.Args[1] + ":" + os.Args[2]
}

//Send response to a client with one property
func sendResponse(w http.ResponseWriter, returnType, property, value string, number int) {
	if returnType == "string" {
		response := map[string]string{property: value}
		responseJSON, _ := json.Marshal(response)
		w.Write(responseJSON)
		return
	}
	if returnType == "int" {
		response := map[string]int{property: number}
		responseJSON, _ := json.Marshal(response)
		w.Write(responseJSON)
		return
	}
	return
}

// Handlers
func incrementValue(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if len(key) < 1 {
		sendResponse(w, "string", error, error600, 0)
		return
	}
	value := r.URL.Query().Get("value")
	if len(value) < 1 {
		sendResponse(w, "string", error, error601, 0)
		return
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     getRedisAddress(),
		Password: password, // no password set
		DB:       DB,       // use default DB
	})
	result := redisClient.Do("INCRBYFLOAT", key, value)
	if err := result.Err(); err != nil {
		sendResponse(w, "string", error, error608, 0)
		return
	}
	incrementedValue, err := strconv.Atoi(result.Val().(string))
	if err != nil {
		sendResponse(w, "string", error, error607, 0)
		return
	}
	sendResponse(w, "int", "value", "", incrementedValue)
}

func getSignature(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if len(text) < 4 {
		sendResponse(w, "string", error, error602, 0)
		return
	}
	key := r.URL.Query().Get("key")
	if len(key) < 4 {
		sendResponse(w, "string", error, error600, 0)
		return
	}
	// Create a new HMAC by defining the hash type and the key (as byte array)
	hmacWriter := hmac.New(sha256.New, []byte(key))
	// Write Data to it
	hmacWriter.Write([]byte(text))
	// Get result and encode as hexadecimal string
	hexSignature := hex.EncodeToString(hmacWriter.Sum(nil))
	sendResponse(w, "string", "signature", hexSignature, 0)
}

func insertUser(w http.ResponseWriter, r *http.Request) {
	connStr := connectionString
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		sendResponse(w, "string", error, error606, 0)
		return
	}
	defer db.Close()
	name := r.URL.Query().Get("name")
	if len(name) < 2 {
		sendResponse(w, "string", error, error604, 0)
		return
	}
	age := r.URL.Query().Get("age")
	if len(age) < 1 {
		sendResponse(w, "string", error, error603, 0)
		return
	}
	if _, err = db.Query("select * from \"users\" limit 1;"); err != nil {
		_, err = db.Exec("CREATE TABLE users (user_id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY, name  varchar(30) NOT NULL, age integer NOT NULL);")
	}
	if err != nil {
		sendResponse(w, "string", error, error605, 0)
		return
	}
	var id int
	db.QueryRow("insert into \"users\" (name, age) values ($1, $2) RETURNING user_id", name, age).Scan(&id)
	sendResponse(w, "int", "id", "", id)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/redis/incr", incrementValue).Methods("POST")
	router.HandleFunc("/sign/hmacsha512", getSignature).Methods("POST")
	router.HandleFunc("/postgres/users", insertUser).Methods("POST")
	http.ListenAndServe(serverAddress, router)
}