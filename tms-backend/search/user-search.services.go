package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type UsersSearchService struct {
	client *elasticsearch.Client
	index  string
}

func NewUsersSearchService() *UsersSearchService {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	return &UsersSearchService{
		client: client,
		index:  "users",
	}
}

// Initialize creates the index with proper mappings if it doesn't exist
func (s *UsersSearchService) Initialize(ctx context.Context) error {
	// Check if index exists
	res, err := s.client.Indices.Exists([]string{s.index})
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		// Index doesn't exist, create it
		log.Println("Need to create index")

		// Define index mappings
		mappings := map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"suggestFullName": map[string]interface{}{
						"type":     "completion",
						"analyzer": "simple",
					},
					"suggestFirstName": map[string]interface{}{
						"type":     "completion",
						"analyzer": "simple",
					},
					"suggestLastName": map[string]interface{}{
						"type":     "completion",
						"analyzer": "simple",
					},
					"suggestFullNameWithWeights": map[string]interface{}{
						"type":     "completion",
						"analyzer": "simple",
					},
					"id": map[string]interface{}{
						"type": "integer",
					},
					"firstName": map[string]interface{}{
						"type": "text",
					},
					"lastName": map[string]interface{}{
						"type": "text",
					},
					"homeAddress": map[string]interface{}{
						"type": "text",
					},
					"landlord": map[string]interface{}{
						"type": "boolean",
					},
				},
			},
		}

		// Convert mappings to JSON
		mappingsJSON, err := json.Marshal(mappings)
		if err != nil {
			return fmt.Errorf("failed to marshal mappings: %w", err)
		}

		// Create index with mappings
		req := esapi.IndicesCreateRequest{
			Index: s.index,
			Body:  bytes.NewReader(mappingsJSON),
		}

		res, err = req.Do(ctx, s.client)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("error creating index: %s", res.String())
		}

		log.Println("Index created successfully")
	}

	return nil
}

func (s *UsersSearchService) IndexUser(ctx context.Context, user models.User) error {
	//Prep doc for indexing
	doc := UserSearchBody{
		ID:          int(user.ID),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		HomeAddress: user.HomeAddress,
		Landlord:    user.Landlord,
		SuggestFullName: map[string]interface{}{
			"input": []string{fmt.Sprintf("%s %s", user.FirstName, user.LastName)},
		},
		SuggestFirstName: map[string]interface{}{
			"input": []string{user.FirstName},
		},
		SuggestLastName: map[string]interface{}{
			"input": []string{user.LastName},
		},
		SuggestFullNameWithWeights: []map[string]interface{}{
			{
				"input":  user.FirstName,
				"weight": 2,
			},
			{
				"input":  user.LastName,
				"weight": 1,
			},
		},
	}

	// Convert doc to JSON
	docJSON, err := json.Marshal(doc)
	if err != nil {
		log.Fatalf("failed to marshal user document: %v", err)
	}

	// Index the document
	req := esapi.IndexRequest{
		Index:      s.index,
		DocumentID: fmt.Sprintf("%d", user.ID),
		Body:       bytes.NewReader(docJSON),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to index user: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing user: %s", res.String())
	}

	return nil
}

// Search performs a multi-match search
func (s *UsersSearchService) Search(ctx context.Context, queryText string) ([]UserSearchBody, error) {
	// prepare query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":    queryText,
				"fields":   []string{"firstName", "lastName"},
				"fuzziness": "AUTO",
			},
		},
	}

	// Convert query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Perform search
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching: %s", res.String())
	}

	// Parse response
	var result UserSearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Extract hits
	var users []UserSearchBody
	for _, hit := range result.Hits.Hits {
		users = append(users, hit.Source)
	}

	return users, nil

}

//suggest returns suggestions for users based on the input text
func (s *UsersSearchService) Suggest(ctx context.Context, queryText string) ([]SuggestionResponse, error) {
	//prep query
	query := map[string]interface{}{
		"suggest": map[string]interface{}{
			"suggestFullName": map[string]interface{}{
				"prefix": queryText,
				"completion": map[string]interface{}{
					"field": "suggestFullName",
					"size": 5,
					"skip_duplicates": false,
					"fuzzy": map[string]interface{}{
						"fuzziness": "AUTO",
					},
					"analyzer": "simple",
				},
			},
			"suggestFirstName": map[string]interface{}{
				"prefix": queryText,
				"completion": map[string]interface{}{
					"field":           "suggestFirstName",
					"size":            5,
					"skip_duplicates": false,
					"fuzzy": map[string]interface{}{
						"fuzziness": "AUTO",
					},
					"analyzer": "simple",
				},
			},
			"suggestLastName": map[string]interface{}{
				"prefix": queryText,
				"completion": map[string]interface{}{
					"field":           "suggestLastName",
					"size":            5,
					"skip_duplicates": false,
					"fuzzy": map[string]interface{}{
						"fuzziness": "AUTO",
					},
					"analyzer": "simple",
				},
			},
		},
	}

	// Convert query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Perform search
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error suggesting: %s", res.String())
	}

	// Parse response
	var result UserSearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract suggestions
	suggestions := make([]SuggestionResponse, 0)
	suggestionIDs := make(map[int]bool)

	// Process first name suggestions
	if result.Suggest != nil && len(result.Suggest.SuggestFirstName) > 0 {
		for _, option := range result.Suggest.SuggestFirstName[0].Options {
			id := option.Source.ID
			if !suggestionIDs[id] {
				suggestionIDs[id] = true
				suggestions = append(suggestions, SuggestionResponse{
					Score:       option.Score,
					ID:          id,
					FirstName:   option.Source.FirstName,
					LastName:    option.Source.LastName,
					HomeAddress: option.Source.HomeAddress,
					Landlord:    option.Source.Landlord,
				})
			}
		}
	}

	// Process last name suggestions
	if result.Suggest != nil && len(result.Suggest.SuggestLastName) > 0 {
		for _, option := range result.Suggest.SuggestLastName[0].Options {
			id := option.Source.ID
			if !suggestionIDs[id] {
				suggestionIDs[id] = true
				suggestions = append(suggestions, SuggestionResponse{
					Score:       option.Score,
					ID:          id,
					FirstName:   option.Source.FirstName,
					LastName:    option.Source.LastName,
					HomeAddress: option.Source.HomeAddress,
					Landlord:    option.Source.Landlord,
				})
			}
		}
	}

	// Process full name suggestions
	if result.Suggest != nil && len(result.Suggest.SuggestFullName) > 0 {
		for _, option := range result.Suggest.SuggestFullName[0].Options {
			id := option.Source.ID
			if !suggestionIDs[id] {
				suggestionIDs[id] = true
				suggestions = append(suggestions, SuggestionResponse{
					Score:       option.Score,
					ID:          id,
					FirstName:   option.Source.FirstName,
					LastName:    option.Source.LastName,
					HomeAddress: option.Source.HomeAddress,
					Landlord:    option.Source.Landlord,
				})
			}
		}
	}

	return suggestions, nil
}

// Remove deletes a user from the index by ID
func (s *UsersSearchService) Remove(ctx context.Context, userID int) error {
	// Prepare delete query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"id": userID,
			},
		},
	}

	// Convert query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	// Execute delete by query
	refresh := true
	req := esapi.DeleteByQueryRequest{
		Index:   []string{s.index},
		Body:    bytes.NewReader(queryJSON),
		Refresh: &refresh,
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting: %s", res.String())
	}

	return nil
}

// Update updates a user in the index
func (s *UsersSearchService) Update(ctx context.Context, user models.User) error {
	// Remove the user first
	if err := s.Remove(ctx, int(user.ID)); err != nil {
		return fmt.Errorf("failed to remove user for update: %w", err)
	}

	// Re-index the user
	if err := s.IndexUser(ctx, user); err != nil {
		return fmt.Errorf("failed to re-index user for update: %w", err)
	}

	return nil
}

// SearchWithSuggest performs a combined search and suggest operation
func (s *UsersSearchService) SearchWithSuggest(ctx context.Context, queryText string) ([]UserSearchBody, error) {
	// Prepare suggest query
	query := map[string]interface{}{
		"suggest": map[string]interface{}{
			"suggestFullName": map[string]interface{}{
				"prefix": queryText,
				"completion": map[string]interface{}{
					"field":           "suggestFullName",
					"size":            5,
					"skip_duplicates": false,
					"fuzzy": map[string]interface{}{
						"fuzziness": 1,
					},
					"analyzer": "simple",
				},
			},
		},
	}

	// Convert query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Execute search with suggest
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search with suggest: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching with suggest: %s", res.String())
	}

	// Parse response
	var result UserSearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract sources from hits
	var users []UserSearchBody
	for _, hit := range result.Hits.Hits {
		users = append(users, hit.Source)
	}

	return users, nil
}