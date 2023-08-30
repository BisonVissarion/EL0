package application

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
		w.Write([]byte("Missing 'id' parameter"))
		return
	}

	if cachedData, found := dataCache.Get(id); found {
		if contact, ok := cachedData.(*Contact); ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(contact)
			return
		}
	}

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

		nc, err := nats.Connect("nats://localhost:4222")
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
