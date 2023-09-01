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
	"strconv"
	"strings"
	"time"

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
	receivedContacts := []*Contact{
		{
			ID:        1,
			Name:      "John Doe",
			Address:   "123 Main St, City",
			Phone:     "555-123-4567",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "Jane Smith",
			Address:   "456 Elm St, Town",
			Phone:     "555-987-6543",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	flag.Parse()

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Невозможно подключиться к NATS: %v\n", err)
	}
	defer nc.Close()

	// Отправляем контакты в NATS
	sendContactsToNATS(nc)

	// Ожидание некоторое время, чтобы код не завершился сразу
	time.Sleep(5 * time.Second)

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

	nc, err = nats.Connect("nats://localhost:4222")
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
	for _, contact := range receivedContacts {
		query := "INSERT INTO contacts (id, name, address, created_at, updated_at, phone) VALUES ($1, $2, $3, $4, $5, $6)"
		_, err := db.Exec(query, strconv.Itoa(contact.ID), contact.Name, contact.Address, contact.CreatedAt, contact.UpdatedAt, contact.Phone)
		if err != nil {
			log.Printf("Ошибка при вставке контакта в базу данных: %v\n", err)
		} else {
			log.Println("Контакт успешно сохранен в базе данных.")
		}

	}

	// Настройка HTTP сервера
	http.HandleFunc("/", handler)
	log.Printf("Слушаем на %s\n", *listenAddr)
	http.ListenAndServe(*listenAddr, nil)

}
