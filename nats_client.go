package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func sendContactsToNATS(nc *nats.Conn) {
	// Создаем контакты для отправки в NATS
	contacts := []*Contact{
		{
			ID:        1,
			Name:      "John Doe",
			Address:   "123 Main St, City",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Phone:     "555-123-4567",
			Favorites: &ContactFavorites{
				Colors: []string{"Blue", "Green"},
			},
		},
		{
			ID:        2,
			Name:      "Jane Smith",
			Address:   "456 Elm St, Town",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Phone:     "555-987-6543",
			Favorites: &ContactFavorites{
				Colors: []string{"Red", "Yellow"},
			},
		},
	}

	// Преобразование данных в JSON
	contactJSON, err := json.Marshal(contacts)
	if err != nil {
		log.Printf("Ошибка при маршалинге контактов в JSON: %v\n", err)
		return
	}

	// Отправка данных в NATS в канал "contacts"
	err = nc.Publish("contacts", contactJSON)
	if err != nil {
		log.Printf("Ошибка при публикации контактов в NATS: %v\n", err)
		return
	}

	log.Println("Контакты успешно отправлены в NATS")
}
