package services

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

type UserSearchService struct {
	client *elasticsearch.Client
	index string
}

func NewUserSearchService() *UserSearchService {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	return &UserSearchService{
		client: client,
		index: "users",
	}
}


// CreateIndex creates the Elasticsearch index with mappings
func (s *UserSearchService) CreateIndex() error {
    mappings := map[string]interface{}{
        "mappings": map[string]interface{}{
            "properties": map[string]interface{}{
                "suggestFullName": map[string]interface{}{
                    "type":     "completion",
                    "analyzer": "simple",
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

    body, _ := json.Marshal(mappings)
    res, err := s.client.Indices.Create(s.index, s.client.Indices.Create.WithBody(bytes.NewReader(body)))
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return fmt.Errorf("error creating index: %s", res.String())
    }
    return nil
}



// IndexUser indexes a user document
func (s *UserSearchService) IndexUser(user map[string]interface{}) error {
    body, _ := json.Marshal(user)
    res, err := s.client.Index(s.index, bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer res.Body.Close()

    if res.IsError() {
        return fmt.Errorf("error indexing user: %s", res.String())
    }
    return nil
}


// Search performs a multi-match search
func (s *UserSearchService) Search(queryText string) ([]map[string]interface{}, error) {
    query := map[string]interface{}{
        "query": map[string]interface{}{
            "multi_match": map[string]interface{}{
                "query":    queryText,
                "fields":   []string{"firstName", "lastName"},
                "fuzziness": "AUTO",
            },
        },
    }

    body, _ := json.Marshal(query)
    res, err := s.client.Search(
        s.client.Search.WithIndex(s.index),
        s.client.Search.WithBody(bytes.NewReader(body)),
    )
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    if res.IsError() {
        return nil, fmt.Errorf("error searching: %s", res.String())
    }

    var result map[string]interface{}
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        return nil, err
    }

    hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
    var users []map[string]interface{}
    for _, hit := range hits {
        source := hit.(map[string]interface{})["_source"].(map[string]interface{})
        users = append(users, source)
    }
    return users, nil
}


// Suggest provides autocomplete suggestions
func (s *UserSearchService) Suggest(prefix string) ([]map[string]interface{}, error) {
    query := map[string]interface{}{
        "suggest": map[string]interface{}{
            "user-suggest": map[string]interface{}{
                "prefix": prefix,
                "completion": map[string]interface{}{
                    "field": "suggestFullName",
                    "size":  5,
                    "fuzzy": map[string]interface{}{
                        "fuzziness": "AUTO",
                    },
                },
            },
        },
    }

    body, _ := json.Marshal(query)
    res, err := s.client.Search(
        s.client.Search.WithIndex(s.index),
        s.client.Search.WithBody(bytes.NewReader(body)),
    )
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    if res.IsError() {
        return nil, fmt.Errorf("error suggesting: %s", res.String())
    }

    var result map[string]interface{}
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        return nil, err
    }

    options := result["suggest"].(map[string]interface{})["user-suggest"].([]interface{})[0].(map[string]interface{})["options"].([]interface{})
    var suggestions []map[string]interface{}
    for _, option := range options {
        source := option.(map[string]interface{})["_source"].(map[string]interface{})
        suggestions = append(suggestions, source)
    }
    return suggestions, nil
}

// SearchWithSuggest performs a search with suggestions
func (s *UserSearchService) SearchWithSuggest(queryText string) ([]map[string]interface{}, error) {
    query := map[string]interface{}{
        "query": map[string]interface{}{
            "multi_match": map[string]interface{}{
                "query":    queryText,
                "fields":   []string{"firstName", "lastName"},
                "fuzziness": 2,
            },
        },
        "suggest": map[string]interface{}{
            "suggestFullName": map[string]interface{}{
                "prefix": queryText,
                "completion": map[string]interface{}{
                    "field": "suggestFullName",
                    "size":  5,
                    "skip_duplicates": false,
                    "fuzzy": map[string]interface{}{
                        "fuzziness": 1,
                    },
                    "analyzer": "simple",
                },
            },
        },
    }

    body, _ := json.Marshal(query)
    res, err := s.client.Search(
        s.client.Search.WithIndex(s.index),
        s.client.Search.WithBody(bytes.NewReader(body)),
    )
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    if res.IsError() {
        return nil, fmt.Errorf("error searching with suggest: %s", res.String())
    }

    var result map[string]interface{}
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        return nil, err
    }

    // Extract hits
    hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
    var users []map[string]interface{}
    for _, hit := range hits {
        source := hit.(map[string]interface{})["_source"].(map[string]interface{})
        users = append(users, source)
    }

    // Extract suggestions
    suggestions := result["suggest"].(map[string]interface{})["suggestFullName"].([]interface{})[0].(map[string]interface{})["options"].([]interface{})
    for _, suggestion := range suggestions {
        source := suggestion.(map[string]interface{})["_source"].(map[string]interface{})
        users = append(users, source)
    }

    return users, nil
}

// Remove deletes a user document by ID
func (s *UserSearchService) Remove(userID string) error {
    query := map[string]interface{}{
        "query": map[string]interface{}{
            "match": map[string]interface{}{
                "id": userID,
            },
        },
    }

    body, _ := json.Marshal(query)
    res, err := s.client.DeleteByQuery(
        []string{s.index},
        bytes.NewReader(body),
    )
    if err != nil {
        return fmt.Errorf("error deleting user by ID: %v", err)
    }
    defer res.Body.Close()

    if res.IsError() {
        return fmt.Errorf("error deleting user by ID: %s", res.String())
    }
    return nil
}

// Update updates a user document
func (s *UserSearchService) Update(user map[string]interface{}) error {
    userID := user["id"].(string)

    // Remove the existing document
    if err := s.Remove(userID); err != nil {
        return fmt.Errorf("error removing user for update: %v", err)
    }

    // Re-index the updated document
    if err := s.IndexUser(user); err != nil {
        return fmt.Errorf("error re-indexing user: %v", err)
    }

    return nil
}