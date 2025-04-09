// search/interfaces.go
package search

// UserSearchBody represents the structure for indexing a user in Elasticsearch
type UserSearchBody struct {
	ID                        int                       `json:"id"`
	FirstName                 string                    `json:"firstName"`
	LastName                  string                    `json:"lastName"`
	HomeAddress               string                    `json:"homeAddress"`
	Landlord                  bool                      `json:"landlord"`
	SuggestFullName           map[string]interface{}    `json:"suggestFullName,omitempty"`
	SuggestFirstName          map[string]interface{}    `json:"suggestFirstName,omitempty"`
	SuggestLastName           map[string]interface{}    `json:"suggestLastName,omitempty"`
	SuggestFullNameWithWeights []map[string]interface{} `json:"suggestFullNameWithWeights,omitempty"`
}

// SuggestionOption represents a single suggestion option in search results
type SuggestionOption struct {
	Text    string        `json:"text"`
	Score   float64       `json:"_score"`
	Source  UserSearchBody `json:"_source"`
}

// SuggestionResult represents a single suggestion result
type SuggestionResult struct {
	Text    string             `json:"text"`
	Offset  int                `json:"offset"`
	Length  int                `json:"length"`
	Options []SuggestionOption `json:"options"`
}

// UserSuggestionMap contains the different suggestion types
type UserSuggestionMap struct {
	SuggestFullNameWithWeights []SuggestionResult `json:"suggestFullNameWithWeights,omitempty"`
	SuggestFullName           []SuggestionResult `json:"suggestFullName,omitempty"`
	SuggestFirstName          []SuggestionResult `json:"suggestFirstName,omitempty"`
	SuggestLastName           []SuggestionResult `json:"suggestLastName,omitempty"`
}

// SearchHit represents a single hit in the search results
type SearchHit struct {
	Source UserSearchBody `json:"_source"`
}

// SearchHits represents the hits section of search results
type SearchHits struct {
	Total struct {
		Value int `json:"value"`
	} `json:"total"`
	Hits []SearchHit `json:"hits"`
}

// UserSearchResult represents the result of a user search
type UserSearchResult struct {
	Hits    *SearchHits       `json:"hits,omitempty"`
	Suggest *UserSuggestionMap `json:"suggest,omitempty"`
}

// SuggestionResponse represents a simplified suggestion to return to clients
type SuggestionResponse struct {
	Score       float64 `json:"score"`
	ID          int     `json:"id"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	HomeAddress string  `json:"homeAddress"`
	Landlord    bool    `json:"landlord"`
}