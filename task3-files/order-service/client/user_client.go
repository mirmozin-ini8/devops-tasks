package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"order-service/model"
	"os"
)

func GetUser(UserID int) (*model.UserResponse, error) {
	baseURL := os.Getenv("USER_SERVICE_URL")
	url := fmt.Sprintf("%s/users/%d", baseURL, UserID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("user service unreachable: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status: %d", resp.StatusCode)
	}

	var user model.UserResponse

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %w", err)
	}

	return &user, nil
}
