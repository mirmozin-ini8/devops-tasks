package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"order-service/model"
	"os"
)

func GetBook(bookID int) (*model.BookResponse, error) {
	baseURL := os.Getenv("BOOK_SERVICE_URL")
	url := fmt.Sprintf("%s/books/%d", baseURL, bookID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("book service unreachable: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("book service returned status: %d", resp.StatusCode)
	}

	var book model.BookResponse
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return nil, fmt.Errorf("failed to parse book response: %w", err)
	}

	return &book, nil
}
