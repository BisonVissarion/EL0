// config.go

package main

import (
	"flag"
	"html/template"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
)

// ContactFavorites - это структура, которая содержит избранные цвета контакта
type ContactFavorites struct {
	Colors []string `json:"colors"`
}

// Contact представляет модель контакта в базе данных
type Contact struct {
	ID                   int
	Name, Address, Phone string
	FavoritesJSON        types.JSONText    `db:"favorites"`
	Favorites            *ContactFavorites `db:"-"`
	CreatedAt            string            `db:"created_at"`
	UpdatedAt            string            `db:"updated_at"`
}

var (
	connectionString = flag.String("conn", getenvWithDefault("DATABASE_URL", ""), "Строка подключения к PostgreSQL")
	listenAddr       = flag.String("addr", getenvWithDefault("LISTENADDR", ":8080"), "HTTP адрес для прослушивания")
	db               *sqlx.DB
	tmpl             = template.New("")
)
