package interfaces

type UserSearchBody struct {
	ID                         int                       `json:"id"`
	FirstName                  string                    `json:"firstName"`
	LastName                   string                    `json:"lastName"`
	HomeAddress                string                    `json:"homeAddress"`
	Landlord                   *bool                     `json:"landlord,omitempty"`
	SuggestFullName            *map[string]interface{}   `json:"suggestFullName,omitempty"`
	SuggestFirstName           *map[string]interface{}   `json:"suggestFirstName,omitempty"`
	SuggestLastName            *map[string]interface{}   `json:"suggestLastName,omitempty"`
	SuggestFullNameWithWeights *[]map[string]interface{} `json:"suggestFullNameWithWeights"`
}
