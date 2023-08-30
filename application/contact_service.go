package application

import (
	"encoding/json"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

type ContactFavorites struct {
	Colors []string `json:"colors"`
}

type Contact struct {
	ID        int
	Name      string
	Address   string
	Phone     string
	Favorites *ContactFavorites
	CreatedAt string
	UpdatedAt string
}

type ContactService struct {
	DB        *sqlx.DB
	NATSConn  *nats.Conn
	DataCache *cache.Cache
}

func NewContactService(db *sqlx.DB, nc *nats.Conn, cache *cache.Cache) *ContactService {
	return &ContactService{
		DB:        db,
		NATSConn:  nc,
		DataCache: cache,
	}
}

func (cs *ContactService) FetchContacts() ([]*Contact, error) {
	contacts := []*Contact{}

	if cachedContacts, found := cs.DataCache.Get("contacts"); found {
		if cachedContacts, ok := cachedContacts.([]*Contact); ok {
			return cachedContacts, nil
		}
	}

	err := cs.DB.Select(&contacts, "SELECT * FROM contacts")
	if err != nil {
		return nil, errors.Wrap(err, "Unable to fetch contacts")
	}

	for _, contact := range contacts {
		err := json.Unmarshal(contact.FavoritesJSON, &contact.Favorites)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to parse JSON favorites")
		}
	}

	cs.DataCache.Set("contacts", contacts, cache.NoExpiration)

	return contacts, nil
}

func (cs *ContactService) GetContactByID(id int) (*Contact, bool) {
	contacts, err := cs.FetchContacts()
	if err != nil {
		log.Printf("Error fetching contacts: %v", err)
		return nil, false
	}

	for _, c := range contacts {
		if c.ID == id {
			return c, true
		}
	}

	return nil, false
}
