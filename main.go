package main

import (
	// Импортируйте ваши собственные пакеты, если они используются

	// Импортируйте другие стандартные или сторонние зависимости, если необходимо
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Импортируйте зависимости для базы данных и NATS, если используете
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
)

var (
	connectionString = flag.String("conn", getenvWithDefault("DATABASE_URL", ""), "PostgreSQL connection string")
	listenAddr       = flag.String("addr", getenvWithDefault("LISTENADDR", ":8080"), "HTTP address to listen on")
	db               *sqlx.DB
	tmpl             = template.New("")
)

func getenvWithDefault(name, defaultValue string) string {
	val := os.Getenv(name)
	if val == "" {
		val = defaultValue
	}
	return val
}

var dataCache = cache.New(cache.NoExpiration, cache.NoExpiration)

func main() {
	http.HandleFunc("/contact/view", contactViewHandler)
	http.HandleFunc("/contact", contactHandler)

	flag.Parse()
	var err error

	// Templating
	tmpl.Funcs(template.FuncMap{"StringsJoin": strings.Join})
	_, err = tmpl.ParseGlob(filepath.Join(".", "templates", "*.html"))
	if err != nil {
		log.Fatalf("Unable to parse templates: %v\n", err)
	}

	// PostgreSQL connection
	if *connectionString == "" {
		log.Fatalln("Please pass the connection string using the -conn option")
	}
	db, err = sqlx.Connect("pgx", *connectionString)
	if err != nil {
		log.Fatalf("Unable to establish connection: %v\n", err)
	}

	// NATS connection
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Unable to connect to NATS: %v\n", err)
	}
	defer nc.Close()

	// NATS subscription
	_, err = nc.Subscribe("contacts", func(msg *nats.Msg) {
		var receivedContacts []*Contact
		err := json.Unmarshal(msg.Data, &receivedContacts)
		if err != nil {
			log.Printf("Error unmarshaling contacts: %v\n", err)
			return
		}
		fmt.Printf("Received Contacts: %+v\n", receivedContacts)
	})
	if err != nil {
		log.Printf("Error subscribing: %v\n", err)
	}

	// HTTP server
	http.HandleFunc("/", handler)
	log.Printf("listening on %s\n", *listenAddr)
	http.ListenAndServe(*listenAddr, nil)
}
