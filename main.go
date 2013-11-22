package main

import (
	"flag"
	"fmt"
	"github.com/trevex/golem"
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
	"net/http"
)

const (
        maxConnections = 5
        connectTimeout = time.Duration(10) * time.Second
        readTimeout = time.Duration(10) * time.Second
        writeTimeout = time.Duration(10) * time.Second
        server = "localhost:6379"
)

var redisPool = redis.NewPool(func() (redis.Conn, error) {
        fmt.Printf("redis.NewPool")
		c, err := redis.DialTimeout("tcp", server, connectTimeout, readTimeout, writeTimeout)
        if err != nil {
			fmt.Printf("redis.NewPool err=%s\n", err)
            return nil, err
        }

        return c, err
}, maxConnections)

var addr = flag.String("addr", ":8080", "http service address")

// This struct represents the message which is accepted by
// the hello-function.
// If a handler function takes a special data
// type that is not an byte array, golem automatically
// tries to unmarshal the data into the specific type.
type RedisKey struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

func redisget(conn *golem.Connection, key *RedisKey) {
	redisConn := redisPool.Get()
    defer redisConn.Close()
    reply, err := redis.String(redisConn.Do("GET", key.Key))
    
    if err != nil {
		fmt.Printf("redisget key=%s err=%s\n", key.Key, err)
		reply = "Key not found."
    }
	conn.Emit("redis_response", &RedisKey{Key:key.Key, Value: reply})
}

func main() {
	flag.Parse()
	
	fmt.Println("http server running on 8080...")

	// Create a router
	myrouter := golem.NewRouter()
	// Add the events to the router
	myrouter.On("redisget", redisget)

	// Serve the public files
	http.Handle("/", http.FileServer(http.Dir("./public")))
	// Handle websockets using golems handler
	http.HandleFunc("/ws", myrouter.Handler())

	// Listen
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
