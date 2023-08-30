package domain

type Contact struct {
	ID        int
	Name      string
	Address   string
	Phone     string
	Favorites *ContactFavorites
}

type ContactFavorites struct {
	Colors []string `json:"colors"`
}
