package utils

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type groupInfo struct {
    ID        string `json:"id"`
    InternalID string `json:"internal_id"`
}

func FindSendingId(groupId string) (string, error) {
    phone := os.Getenv("PHONE_NUMBER")
	serverUrl := os.Getenv("SERVER_URL")
    if phone == "" {
        return "", fmt.Errorf("PHONE_NUMBER is not set")
    }

    resp, err := http.Get(serverUrl + "/v1/groups/" + phone)
    if err != nil {
        return "", fmt.Errorf("fetch groups: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return "", fmt.Errorf("fetch groups failed: %s", resp.Status)
    }

    var groups []groupInfo
    if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
        return "", fmt.Errorf("decode groups response: %w", err)
    }

    for _, g := range groups {
        if g.InternalID == groupId {
            return g.ID, nil
        }
    }

    return "", fmt.Errorf("no sending id found for group %q", groupId)
}