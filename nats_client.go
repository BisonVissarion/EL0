package main

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

func SendDataToNATS() {
	// Создаем соединение с сервером NATS Streaming
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Unable to connect to NATS: %v\n", err)
	}
	defer nc.Close()

	// Отправляем данные в NATS Streaming
	contacts := []Contact{ // Убран лишний тип *Contact
		{
			ID:      1,
			Name:    "John Doe",
			Address: "123 Main St, City",
			Phone:   "555-123-4567",
			Favorites: &ContactFavorites{
				Colors: []string{"Blue", "Green"},
			},
		},
		{
			ID:      2,
			Name:    "Jane Smith",
			Address: "456 Elm St, Town",
			Phone:   "555-987-6543",
			Favorites: &ContactFavorites{
				Colors: []string{"Red", "Yellow"},
			},
		},
	}

	contactJSON, err := json.Marshal(contacts)
	if err != nil {
		log.Fatalf("Unable to marshal contacts to JSON: %v\n", err)
	}

	err = nc.Publish("contacts", contactJSON)
	if err != nil {
		log.Fatalf("Unable to publish contacts: %v\n", err)
	}
}
