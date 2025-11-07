package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getWorkspaceID(apiKey string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.clockify.me/api/v1/workspaces", nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var workspaces []struct {
		ID string `json:"id"`
	}
	json.Unmarshal(body, &workspaces)
	if len(workspaces) == 0 {
		return "", fmt.Errorf("no workspaces found")
	}
	return workspaces[0].ID, nil // Use the first workspace
}

func getUserID(apiKey string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.clockify.me/api/v1/user", nil)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var user struct {
		ID string `json:"id"`
	}
	json.Unmarshal(body, &user)
	return user.ID, nil
}
