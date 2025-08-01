package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joho/godotenv"
)

// Constants
const (
	APIBase               = "https://api.controld.com/profiles"
	BatchSize             = 500
	MaxRetries            = 3
	RetryDelay            = 1 * time.Second
	FolderCreationDelay   = 2 * time.Second
	HTTPTimeout           = 30 * time.Second
	MaxConcurrentProfiles = 3 // Maximum number of profiles to sync concurrently
)

var FolderURLs = []string{
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/apple-private-relay-allow-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-amazon-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-apple-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-huawei-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/ultimate-known_issues-allow-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-lgwebos-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-microsoft-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-oppo-realme-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/referral-allow-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-roku-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-samsung-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/spam-idns-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/spam-tlds-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/spam-tlds-allow-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-tiktok-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-vivo-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/native-tracker-xiaomi-folder.json",
	"https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/badware-hoster-folder.json",
}

// Structs for JSON data
type Action struct {
	Do     int `json:"do"`
	Status int `json:"status"`
}

type Group struct {
	Group  string `json:"group"`
	Action Action `json:"action"`
}

type Rule struct {
	PK string `json:"PK"`
}

type FolderData struct {
	Group Group  `json:"group"`
	Rules []Rule `json:"rules"`
}

type APIGroup struct {
	Group string      `json:"group"`
	PK    interface{} `json:"PK"`
}

type APIGroupsResponse struct {
	Body struct {
		Groups []APIGroup `json:"groups"`
	} `json:"body"`
}

type APIRulesResponse struct {
	Body struct {
		Rules []Rule `json:"rules"`
	} `json:"body"`
}

// Global variables
var (
	token      string
	profileIDs []string
	apiClient  *http.Client
	ghClient   *http.Client
	cache      = make(map[string]FolderData)
)

// Logger setup
func setupLogger() {
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("")
}

// Initialize HTTP clients
func initClients() {
	apiClient = &http.Client{
		Timeout: HTTPTimeout,
	}
	ghClient = &http.Client{
		Timeout: HTTPTimeout,
	}
}

// Retry request with exponential backoff
func retryRequest(requestFunc func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		resp, err := requestFunc()
		if err == nil && resp.StatusCode < 400 {
			return resp, nil
		}

		lastErr = err
		if resp != nil && resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		if attempt == MaxRetries-1 {
			break
		}

		waitTime := RetryDelay * time.Duration(1<<attempt)
		log.Printf("Request failed (attempt %d/%d): %v. Retrying in %v...", attempt+1, MaxRetries, lastErr, waitTime)
		time.Sleep(waitTime)
	}

	return nil, lastErr
}

// API GET request
func apiGet(endpoint string) (*http.Response, error) {
	return retryRequest(func() (*http.Response, error) {
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		return apiClient.Do(req)
	})
}

// API DELETE request
func apiDelete(endpoint string) (*http.Response, error) {
	return retryRequest(func() (*http.Response, error) {
		req, err := http.NewRequest("DELETE", endpoint, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token)

		return apiClient.Do(req)
	})
}

// API POST request
func apiPost(endpoint string, data map[string]string) (*http.Response, error) {
	return retryRequest(func() (*http.Response, error) {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		return apiClient.Do(req)
	})
}

// API POST form request
func apiPostForm(endpoint string, data map[string]string) (*http.Response, error) {
	return retryRequest(func() (*http.Response, error) {
		formData := url.Values{}
		for k, v := range data {
			formData.Set(k, v)
		}

		req, err := http.NewRequest("POST", endpoint, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		return apiClient.Do(req)
	})
}

// GitHub GET request (cached)
func ghGet(url string) (FolderData, error) {
	if data, exists := cache[url]; exists {
		return data, nil
	}

	resp, err := ghClient.Get(url)
	if err != nil {
		return FolderData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return FolderData{}, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var data FolderData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return FolderData{}, err
	}

	cache[url] = data
	return data, nil
}

// Convert interface{} to string
func interfaceToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', 0, 64)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// List existing folders
func listExistingFolders(profileID string) (map[string]string, error) {
	endpoint := fmt.Sprintf("%s/%s/groups", APIBase, profileID)
	resp, err := apiGet(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing folders: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode groups response: %w", err)
	}

	folders := make(map[string]string)
	for _, folder := range apiResp.Body.Groups {
		pkStr := interfaceToString(folder.PK)
		if folder.Group != "" && pkStr != "" {
			folders[strings.TrimSpace(folder.Group)] = pkStr
		}
	}

	return folders, nil
}

// Get all existing rules
func getAllExistingRules(profileID string) (map[string]bool, error) {
	allRules := make(map[string]bool)

	// Get rules from root folder
	endpoint := fmt.Sprintf("%s/%s/rules", APIBase, profileID)
	resp, err := apiGet(endpoint)
	if err != nil {
		log.Printf("Warning: Failed to get root folder rules: %v", err)
	} else {
		defer resp.Body.Close()
		var apiResp APIRulesResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err == nil {
			for _, rule := range apiResp.Body.Rules {
				if rule.PK != "" {
					allRules[rule.PK] = true
				}
			}
			log.Printf("Found %d rules in root folder", len(apiResp.Body.Rules))
		}
	}

	// Get all folders
	folders, err := listExistingFolders(profileID)
	if err != nil {
		return allRules, err
	}

	// Get rules from each folder
	for folderName, folderID := range folders {
		endpoint := fmt.Sprintf("%s/%s/rules/%s", APIBase, profileID, folderID)
		resp, err := apiGet(endpoint)
		if err != nil {
			log.Printf("Warning: Failed to get rules from folder '%s': %v", folderName, err)
			continue
		}

		var apiResp APIRulesResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			log.Printf("Warning: Failed to decode rules from folder '%s': %v", folderName, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for _, rule := range apiResp.Body.Rules {
			if rule.PK != "" {
				allRules[rule.PK] = true
			}
		}

		log.Printf("Found %d rules in folder '%s'", len(apiResp.Body.Rules), folderName)
	}

	log.Printf("Total existing rules across all folders: %d", len(allRules))
	return allRules, nil
}

// Fetch folder data from GitHub
func fetchFolderData(url string) (FolderData, error) {
	return ghGet(url)
}

// Delete folder
func deleteFolder(profileID, name, folderID string) bool {
	endpoint := fmt.Sprintf("%s/%s/groups/%s", APIBase, profileID, folderID)
	_, err := apiDelete(endpoint)
	if err != nil {
		log.Printf("Failed to delete folder '%s' (ID %s): %v", name, folderID, err)
		return false
	}

	log.Printf("Deleted folder '%s' (ID %s)", name, folderID)
	return true
}

// Create folder
func createFolder(profileID, name string, do, status int) (string, error) {
	endpoint := fmt.Sprintf("%s/%s/groups", APIBase, profileID)
	data := map[string]string{
		"name":   name,
		"do":     strconv.Itoa(do),
		"status": strconv.Itoa(status),
	}

	_, err := apiPost(endpoint, data)
	if err != nil {
		return "", fmt.Errorf("failed to create folder '%s': %w", name, err)
	}

	// Re-fetch the list and find the folder we just created
	folders, err := listExistingFolders(profileID)
	if err != nil {
		return "", fmt.Errorf("failed to list folders after creation: %w", err)
	}

	folderID, exists := folders[strings.TrimSpace(name)]
	if !exists {
		return "", fmt.Errorf("folder '%s' was not found after creation", name)
	}

	log.Printf("Created folder '%s' (ID %s)", name, folderID)
	time.Sleep(FolderCreationDelay)
	return folderID, nil
}

// Push rules in batches
func pushRules(profileID, folderName, folderID string, do, status int, hostnames []string, existingRules map[string]bool) bool {
	if len(hostnames) == 0 {
		log.Printf("Folder '%s' - no rules to push", folderName)
		return true
	}

	// Filter out duplicates
	originalCount := len(hostnames)
	var filteredHostnames []string
	for _, hostname := range hostnames {
		if !existingRules[hostname] {
			filteredHostnames = append(filteredHostnames, hostname)
		}
	}

	duplicatesCount := originalCount - len(filteredHostnames)
	if duplicatesCount > 0 {
		log.Printf("Folder '%s': skipping %d duplicate rules", folderName, duplicatesCount)
	}

	if len(filteredHostnames) == 0 {
		log.Printf("Folder '%s' - no new rules to push after filtering duplicates", folderName)
		return true
	}

	successfulBatches := 0
	totalBatches := (len(filteredHostnames) + BatchSize - 1) / BatchSize

	for i := 0; i < len(filteredHostnames); i += BatchSize {
		end := i + BatchSize
		if end > len(filteredHostnames) {
			end = len(filteredHostnames)
		}
		batch := filteredHostnames[i:end]
		batchNum := (i / BatchSize) + 1

		data := map[string]string{
			"do":     strconv.Itoa(do),
			"status": strconv.Itoa(status),
			"group":  folderID,
		}

		for j, hostname := range batch {
			data[fmt.Sprintf("hostnames[%d]", j)] = hostname
		}

		endpoint := fmt.Sprintf("%s/%s/rules", APIBase, profileID)
		_, err := apiPostForm(endpoint, data)
		if err != nil {
			log.Printf("Failed to push batch %d for folder '%s': %v", batchNum, folderName, err)
			continue
		}

		log.Printf("Folder '%s' – batch %d: added %d rules", folderName, batchNum, len(batch))
		successfulBatches++

		// Update existing rules set
		for _, hostname := range batch {
			existingRules[hostname] = true
		}
	}

	if successfulBatches == totalBatches {
		log.Printf("Folder '%s' – finished (%d new rules added)", folderName, len(filteredHostnames))
		return true
	} else {
		log.Printf("Folder '%s' – only %d/%d batches succeeded", folderName, successfulBatches, totalBatches)
		return false
	}
}

// Sync profile
func syncProfile(profileID string) bool {
	log.Printf("Starting sync for profile %s", profileID)

	// Fetch all folder data first
	var folderDataList []FolderData
	for _, url := range FolderURLs {
		folderData, err := fetchFolderData(url)
		if err != nil {
			log.Printf("Failed to fetch folder data from %s: %v", url, err)
			continue
		}
		folderDataList = append(folderDataList, folderData)
	}

	if len(folderDataList) == 0 {
		log.Printf("No valid folder data found")
		return false
	}

	// Get existing folders and delete target folders
	existingFolders, err := listExistingFolders(profileID)
	if err != nil {
		log.Printf("Failed to list existing folders: %v", err)
		return false
	}

	for _, folderData := range folderDataList {
		name := strings.TrimSpace(folderData.Group.Group)
		if folderID, exists := existingFolders[name]; exists {
			deleteFolder(profileID, name, folderID)
		}
	}

	// Get all existing rules AFTER deleting target folders
	existingRules, err := getAllExistingRules(profileID)
	if err != nil {
		log.Printf("Failed to get existing rules: %v", err)
		return false
	}

	// Create new folders and push rules
	successCount := 0
	for _, folderData := range folderDataList {
		name := strings.TrimSpace(folderData.Group.Group)
		do := folderData.Group.Action.Do
		status := folderData.Group.Action.Status

		var hostnames []string
		for _, rule := range folderData.Rules {
			if rule.PK != "" {
				hostnames = append(hostnames, rule.PK)
			}
		}

		folderID, err := createFolder(profileID, name, do, status)
		if err != nil {
			log.Printf("Failed to create folder '%s': %v", name, err)
			continue
		}

		if pushRules(profileID, name, folderID, do, status, hostnames, existingRules) {
			successCount++
		}
	}

	log.Printf("Sync complete: %d/%d folders processed successfully", successCount, len(folderDataList))
	return successCount == len(folderDataList)
}

// Main function
func main() {
	setupLogger()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	token = os.Getenv("TOKEN")
	profilesEnv := os.Getenv("PROFILE")

	if token == "" || profilesEnv == "" {
		log.Fatal("TOKEN and/or PROFILE missing - check your .env file")
	}

	// Parse profile IDs
	for _, p := range strings.Split(profilesEnv, ",") {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			profileIDs = append(profileIDs, trimmed)
		}
	}

	if len(profileIDs) == 0 {
		log.Fatal("No valid profile IDs found")
	}

	initClients()

	// Use goroutines for concurrent profile syncing with semaphore to limit concurrency
	semaphore := make(chan struct{}, MaxConcurrentProfiles)
	var wg sync.WaitGroup
	var successCount int32

	log.Printf("Starting concurrent sync for %d profiles (max %d concurrent)", len(profileIDs), MaxConcurrentProfiles)

	for _, profileID := range profileIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore

			if syncProfile(id) {
				atomic.AddInt32(&successCount, 1)
			}
		}(profileID)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	finalSuccessCount := int(atomic.LoadInt32(&successCount))
	log.Printf("All profiles processed: %d/%d successful", finalSuccessCount, len(profileIDs))

	if finalSuccessCount != len(profileIDs) {
		os.Exit(1)
	}
}
