package infrastructure

import (
	"encoding/json"

	"github.com/BisonVissarion/EL0/domain"
	"github.com/jmoiron/sqlx"
)

type ContactRepository struct {
	DB *sqlx.DB
}

func NewContactRepository(db *sqlx.DB) *ContactRepository {
	return &ContactRepository{DB: db}
}

func (r *ContactRepository) FetchContacts() ([]*domain.Contact, error) {
	// Выполните запрос к базе данных для получения контактов
	var contactRecords []struct {
		ID        int    `db:"id"`
		Name      string `db:"name"`
		Address   string `db:"address"`
		Phone     string `db:"phone"`
		Favorites string `db:"favorites"` // Для FavoritesJSON, как в вашей структуре Contact
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}

	query := `SELECT id, name, address, phone, favorites, created_at, updated_at FROM contacts`
	if err := r.DB.Select(&contactRecords, query); err != nil {
		return nil, err
	}

	// Преобразуйте записи из базы данных в объекты домена
	contacts := make([]*domain.Contact, len(contactRecords))
	for i, record := range contactRecords {
		contact := &domain.Contact{
			ID:        record.ID,
			Name:      record.Name,
			Address:   record.Address,
			Phone:     record.Phone,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		}

		// Распакуйте JSON для Favorites
		err := json.Unmarshal([]byte(record.Favorites), &contact.Favorites)
		if err != nil {
			return nil, err
		}

		contacts[i] = contact
	}

	return contacts, nil
}

func (r *ContactRepository) GetContactByID(id int) (*domain.Contact, bool) {
	// Выполните запрос к базе данных для получения контакта по ID
	var contactRecord struct {
		ID        int    `db:"id"`
		Name      string `db:"name"`
		Address   string `db:"address"`
		Phone     string `db:"phone"`
		Favorites string `db:"favorites"` // Для FavoritesJSON, как в вашей структуре Contact
		CreatedAt string `db:"created_at"`
		UpdatedAt string `db:"updated_at"`
	}

	query := `SELECT id, name, address, phone, favorites, created_at, updated_at FROM contacts WHERE id = $1`
	if err := r.DB.Get(&contactRecord, query, id); err != nil {
		return nil, false // Контакт не найден
	}

	// Преобразуйте запись из базы данных в объект домена
	contact := &domain.Contact{
		ID:        contactRecord.ID,
		Name:      contactRecord.Name,
		Address:   contactRecord.Address,
		Phone:     contactRecord.Phone,
		CreatedAt: contactRecord.CreatedAt,
		UpdatedAt: contactRecord.UpdatedAt,
	}

	// Распакуйте JSON для Favorites
	err := json.Unmarshal([]byte(contactRecord.Favorites), &contact.Favorites)
	if err != nil {
		return nil, false // Ошибка при разборе JSON
	}

	return contact, true
}
