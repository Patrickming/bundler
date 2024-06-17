package models

type SimpleStorage struct {
	ID          int64  `json:"id"`
	StoredValue uint64 `json:"storedValue"`
}
