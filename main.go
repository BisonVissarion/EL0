package main

import (
	"github.com/patrickmn/go-cache"

	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/nats-io/nats.go"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
)

// ContactFavorites is a field that contains a contact's favorites
type ContactFavorites struct {
	Colors []string `json:"colors"`
}

// Contact represents a Contact model in the database
type Contact struct {
	ID                   int //ID является идентификатором пользователя. Он будет уникальным для каждой записи в таблице пользователей в базе данных.
	Name, Address, Phone string
	FavoritesJSON        types.JSONText    `db:"favorites"`
	Favorites            *ContactFavorites `db:"-"`
	CreatedAt            string            `db:"created_at"`
	UpdatedAt            string            `db:"updated_at"`
}

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
	// templating
	tmpl.Funcs(template.FuncMap{"StringsJoin": strings.Join})
	_, err = tmpl.ParseGlob(filepath.Join(".", "templates", "*.html"))
	if err != nil {
		log.Fatalf("Unable to parse templates: %v\n", err)
	}
	// postgres connection
	if *connectionString == "" {
		log.Fatalln("Please pass the connection string using the -conn option")
	}
	db, err = sqlx.Connect("pgx", *connectionString)
	if err != nil {
		log.Fatalf("Unable to establish connection: %v\n", err)
	}

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Unable to connect to NATS: %v\n", err) //
	}
	defer nc.Close()

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

	// http server
	http.HandleFunc("/", handler)
	log.Printf("listening on %s\n", *listenAddr)
	http.ListenAndServe(*listenAddr, nil)

}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing 'id' parameter"))
		return
	}

	// Проверяем, есть ли данные в кеше по заданному ID
	if cachedData, found := dataCache.Get(id); found {
		if contact, ok := cachedData.(*Contact); ok {
			// Отправляем данные из кеша клиенту
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(contact)
			return
		}
	}

	// Если данных в кеше нет, можно выполнить логику загрузки из базы и добавления в кеш, как в функции fetchContacts

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Contact not found"))
}

func fetchContacts() ([]*Contact, error) {

	contacts := []*Contact{}

	if cachedContacts, found := dataCache.Get("contacts"); found {
		if contacts, ok := cachedContacts.([]*Contact); ok {
			return contacts, nil
		}
	}

	err := db.Select(&contacts, "select * from contacts")
	if err != nil {
		return nil, errors.Wrap(err, "Unable to fetch contacts")
	}
	for _, contact := range contacts {
		err := json.Unmarshal(contact.FavoritesJSON, &contact.Favorites)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse JSON favorites")
		}

		nc, err := nats.Connect("nats://localhost:4222") //
		if err != nil {
			return nil, errors.Wrap(err, "Unable to connect to NATS")
		}
		defer nc.Close()

		contactJSON, _ := json.Marshal(contacts)
		err = nc.Publish("contacts", contactJSON)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to publish contacts")
		}

	}

	dataCache.Set("contacts", contacts, cache.NoExpiration)

	return contacts, nil
}

func contactViewHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing 'id' parameter"))
		return
	}

	contacts, err := fetchContacts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var contact *Contact
	for _, c := range contacts {
		if fmt.Sprintf("%d", c.ID) == id {
			contact = c
			break
		}
	}

	if contact == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Contact not found"))
		return
	}

	tmpl.ExecuteTemplate(w, "contact.html", struct{ Contact *Contact }{contact})
}

func handler(w http.ResponseWriter, r *http.Request) {
	contacts, err := fetchContacts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	tmpl.ExecuteTemplate(w, "index.html", struct{ Contacts []*Contact }{contacts})
}
