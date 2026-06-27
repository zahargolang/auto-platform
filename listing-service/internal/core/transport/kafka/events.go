package core_kafka

type ListingCreatedEvent struct {
	ID     string `json:"listing_id"`
	UserID string `json:"user_id"`
	Title  string `json:"title"`
	Price  int64  `json:"price"`
	Make   string `json:"make"`
	Model  string `json:"model"`
	Year   int    `json:"year"`
	City   string `json:"city"`
}

type ListingDeletedEvent struct {
	ID     string `json:"listing_id"`
	UserID string `json:"user_id"`
}
