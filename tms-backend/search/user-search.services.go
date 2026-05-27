package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/auditrakkr/tms-fullstack/tms-backend/models"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// UsersSearchService manages Elasticsearch operations for the User domain.
type UsersSearchService struct {
	client *elasticsearch.Client
	index  string
}

// NewUsersSearchService constructs the service using an explicit node URL
// (e.g. from config.AppConfig.Elasticsearch.Node).
// Returns an error if the ES client cannot be created so the caller can
// degrade gracefully instead of panicking.
func NewUsersSearchService(node string) (*UsersSearchService, error) {
	if node == "" {
		node = "http://localhost:9200"
	}
	cfg := elasticsearch.Config{
		Addresses: []string{node},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}
	return &UsersSearchService{client: client, index: "users"}, nil
}

// Initialize creates the users index with completion-field mappings if it does not exist.
func (s *UsersSearchService) Initialize(ctx context.Context) error {
	res, err := s.client.Indices.Exists([]string{s.index})
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 404 {
		return nil // index already exists
	}

	mappings := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":          map[string]any{"type": "integer"},
				"firstName":   map[string]any{"type": "text"},
				"lastName":    map[string]any{"type": "text"},
				"homeAddress": map[string]any{"type": "text"},
				"landlord":    map[string]any{"type": "boolean"},
				"suggestFullName": map[string]any{
					"type":     "completion",
					"analyzer": "simple",
				},
				"suggestFirstName": map[string]any{
					"type":     "completion",
					"analyzer": "simple",
				},
				"suggestLastName": map[string]any{
					"type":     "completion",
					"analyzer": "simple",
				},
				"suggestFullNameWithWeights": map[string]any{
					"type":     "completion",
					"analyzer": "simple",
				},
			},
		},
	}

	body, err := json.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("failed to marshal index mappings: %w", err)
	}

	createRes, err := esapi.IndicesCreateRequest{
		Index: s.index,
		Body:  bytes.NewReader(body),
	}.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("elasticsearch error creating index: %s", createRes.String())
	}

	slog.Info("elasticsearch index created", "index", s.index)
	return nil
}

// IndexUser adds or replaces a user document in the index.
func (s *UsersSearchService) IndexUser(ctx context.Context, user models.User) error {
	doc := UserSearchBody{
		ID:          int(user.ID),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		HomeAddress: user.HomeAddress,
		Landlord:    user.Landlord,
		SuggestFullName: map[string]any{
			"input": []string{fmt.Sprintf("%s %s", user.FirstName, user.LastName)},
		},
		SuggestFirstName: map[string]any{"input": []string{user.FirstName}},
		SuggestLastName:  map[string]any{"input": []string{user.LastName}},
		SuggestFullNameWithWeights: []map[string]any{
			{"input": user.FirstName, "weight": 2},
			{"input": user.LastName, "weight": 1},
		},
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal user document: %w", err)
	}

	res, err := esapi.IndexRequest{
		Index:      s.index,
		DocumentID: fmt.Sprintf("%d", user.ID),
		Body:       bytes.NewReader(docJSON),
		Refresh:    "true",
	}.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to index user: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error indexing user: %s", res.String())
	}
	return nil
}

// Search performs a fuzzy multi-match query across firstName and lastName.
func (s *UsersSearchService) Search(ctx context.Context, queryText string) ([]UserSearchBody, error) {
	query := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":     queryText,
				"fields":    []string{"firstName", "lastName"},
				"fuzziness": "AUTO",
			},
		},
	}

	queryJSON, _ := json.Marshal(query)
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch search error: %s", res.String())
	}

	var result UserSearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var users []UserSearchBody
	if result.Hits != nil {
		for _, hit := range result.Hits.Hits {
			users = append(users, hit.Source)
		}
	}
	return users, nil
}

// Suggest returns completion suggestions for the given prefix.
func (s *UsersSearchService) Suggest(ctx context.Context, queryText string) ([]SuggestionResponse, error) {
	query := map[string]any{
		"suggest": map[string]any{
			"suggestFullName": map[string]any{
				"prefix": queryText,
				"completion": map[string]any{
					"field": "suggestFullName", "size": 5,
					"skip_duplicates": false,
					"fuzzy":           map[string]any{"fuzziness": "AUTO"},
					"analyzer":        "simple",
				},
			},
			"suggestFirstName": map[string]any{
				"prefix": queryText,
				"completion": map[string]any{
					"field": "suggestFirstName", "size": 5,
					"skip_duplicates": false,
					"fuzzy":           map[string]any{"fuzziness": "AUTO"},
					"analyzer":        "simple",
				},
			},
			"suggestLastName": map[string]any{
				"prefix": queryText,
				"completion": map[string]any{
					"field": "suggestLastName", "size": 5,
					"skip_duplicates": false,
					"fuzzy":           map[string]any{"fuzziness": "AUTO"},
					"analyzer":        "simple",
				},
			},
		},
	}

	queryJSON, _ := json.Marshal(query)
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.index),
		s.client.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("suggest request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch suggest error: %s", res.String())
	}

	var result UserSearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode suggest response: %w", err)
	}

	seen := make(map[int]bool)
	var suggestions []SuggestionResponse

	appendOptions := func(opts []SuggestionResult) {
		for _, s := range opts {
			for _, o := range s.Options {
				if !seen[o.Source.ID] {
					seen[o.Source.ID] = true
					suggestions = append(suggestions, SuggestionResponse{
						Score: o.Score, ID: o.Source.ID,
						FirstName: o.Source.FirstName, LastName: o.Source.LastName,
						HomeAddress: o.Source.HomeAddress, Landlord: o.Source.Landlord,
					})
				}
			}
		}
	}

	if result.Suggest != nil {
		appendOptions(result.Suggest.SuggestFirstName)
		appendOptions(result.Suggest.SuggestLastName)
		appendOptions(result.Suggest.SuggestFullName)
	}
	return suggestions, nil
}

// Remove deletes a user from the index by ID using delete-by-query.
func (s *UsersSearchService) Remove(ctx context.Context, userID int) error {
	query := map[string]any{
		"query": map[string]any{"match": map[string]any{"id": userID}},
	}
	queryJSON, _ := json.Marshal(query)

	refresh := true
	res, err := esapi.DeleteByQueryRequest{
		Index:   []string{s.index},
		Body:    bytes.NewReader(queryJSON),
		Refresh: &refresh,
	}.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("delete-by-query failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch delete error: %s", res.String())
	}
	return nil
}

// Update is a remove-then-index operation (Elasticsearch has no partial update
// for completion fields).
func (s *UsersSearchService) Update(ctx context.Context, user models.User) error {
	if err := s.Remove(ctx, int(user.ID)); err != nil {
		return fmt.Errorf("remove before update failed: %w", err)
	}
	return s.IndexUser(ctx, user)
}
