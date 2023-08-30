package application

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

type NATSService struct {
	Connection *nats.Conn
}

func NewNATSService(connection *nats.Conn) *NATSService {
	return &NATSService{Connection: connection}
}

func (ns *NATSService) PublishContacts(contacts []*Contact) error {
	// Преобразуйте список контактов в JSON
	contactsJSON, err := json.Marshal(contacts)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal contacts to JSON")
	}

	// Отправьте данные в NATS
	err = ns.NATSConn.Publish("contacts", contactsJSON)
	if err != nil {
		return errors.Wrap(err, "Unable to publish contacts to NATS")
	}

	return nil
}
