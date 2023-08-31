package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

func contactHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Отсутствует параметр 'id'"))
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
	w.Write([]byte("Контакт не найден"))
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
		return nil, errors.Wrap(err, "Невозможно получить контакты")
	}
	for _, contact := range contacts {
		err := json.Unmarshal(contact.FavoritesJSON, &contact.Favorites)
		if err != nil {
			return nil, errors.Wrap(err, "Невозможно разобрать JSON избранных")
		}

		nc, err := nats.Connect("nats://localhost:4222")
		if err != nil {
			return nil, errors.Wrap(err, "Невозможно подключиться к NATS")
		}
		defer nc.Close()

		contactJSON, _ := json.Marshal(contacts)
		err = nc.Publish("contacts", contactJSON)
		if err != nil {
			return nil, errors.Wrap(err, "Невозможно опубликовать контакты")
		}
	}

	dataCache.Set("contacts", contacts, cache.NoExpiration)

	return contacts, nil
}

func contactViewHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Отсутствует параметр 'id'"))
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
		w.Write([]byte("Контакт не найден"))
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
