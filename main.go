package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
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
	// Настройка шаблонов
	tmpl.Funcs(template.FuncMap{"StringsJoin": strings.Join})
	_, err = tmpl.ParseGlob(filepath.Join(".", "templates", "*.html"))
	if err != nil {
		log.Fatalf("Невозможно разобрать шаблоны: %v\n", err)
	}
	// Подключение к PostgreSQL
	if *connectionString == "" {
		log.Fatalln("Пожалуйста, укажите строку подключения с помощью опции -conn")
	}
	db, err = sqlx.Connect("pgx", *connectionString)
	if err != nil {
		log.Fatalf("Невозможно установить соединение: %v\n", err)
	}

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Невозможно подключиться к NATS: %v\n", err)
	}
	defer nc.Close()

	_, err = nc.Subscribe("contacts", func(msg *nats.Msg) {
		var receivedContacts []*Contact
		err := json.Unmarshal(msg.Data, &receivedContacts)
		if err != nil {
			log.Printf("Ошибка при разборе контактов: %v\n", err)
			return
		}

		fmt.Printf("Получены контакты: %+v\n", receivedContacts)
	})
	if err != nil {
		log.Printf("Ошибка подписки: %v\n", err)
	}

	// Настройка HTTP сервера
	http.HandleFunc("/", handler)
	log.Printf("Слушаем на %s\n", *listenAddr)
	http.ListenAndServe(*listenAddr, nil)
}
