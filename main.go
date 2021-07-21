package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-redis/redis"
	"github.com/scribble-rs/scribble.rs/game"
	"github.com/scribble-rs/scribble.rs/game/store"
	"github.com/scribble-rs/scribble.rs/server"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	port                = kingpin.Flag("http-port", "HTTP API Port").Envar("HTTP_PORT").Default("8080").Int()
	redisAddr           = kingpin.Flag("redis-addr", "Redis Address").Envar("REDIS_ADDR").Default("127.0.0.1:6379").String()
	agoraAppID          = kingpin.Flag("agora-app-id", "Agora App ID").Envar("AGORA_APP_ID").Default("89f97462e28540e68a6a90760c9ca113").String()
	agoraAppCertificate = kingpin.Flag("agora-cert", "Agora Cert").Envar("AGORA_CERT").String()
)

func main() {
	kingpin.Parse()

	server.AgoraAppCertificate = *agoraAppCertificate
	server.AgoraAppID = *agoraAppID

	//Setting the seed in order for the petnames to be random.
	rand.Seed(time.Now().UnixNano())

	redisStore, err := store.NewRedisStore(&redis.Options{
		Addr: *redisAddr,
	})
	if err != nil {
		log.Fatal("Could not connect to Redis", err)
	}
	game.Store = redisStore

	err = server.Serve(fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal("HTTP Server error:", err)
	}
}
