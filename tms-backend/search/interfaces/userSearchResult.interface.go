package interfaces


type UserSearchResult struct {
    Hits *struct {
        Total uint `json:"total"`
        Hits  []struct {
            Source UserSearchBody `json:"_source"`
        } `json:"hits"`
    } `json:"hits"`
    Suggest *struct {
        SuggestFullNameWithWeights []struct {
            Text    string `json:"text"`
            Offset  uint   `json:"offset"`
            Length  uint   `json:"length"`
            Options []struct {
                Text   string         `json:"text"`
                Score  float64        `json:"_score"`
                Source UserSearchBody `json:"_source"`
            } `json:"options"`
        } `json:"suggestFullNameWithWeights,omitempty"`
        SuggestFullName []struct {
            Text    string `json:"text"`
            Offset  uint   `json:"offset"`
            Length  uint   `json:"length"`
            Options []struct {
                Text   string         `json:"text"`
                Score  float64        `json:"_score"`
                Source UserSearchBody `json:"_source"`
            } `json:"options"`
        } `json:"suggestFullName,omitempty"`
        SuggestFirstName []struct {
            Text    string `json:"text"`
            Offset  uint   `json:"offset"`
            Length  uint   `json:"length"`
            Options []struct {
                Text   string         `json:"text"`
                Score  float64        `json:"_score"`
                Source UserSearchBody `json:"_source"`
            } `json:"options"`
        } `json:"suggestFirstName,omitempty"`
        SuggestLastName []struct {
            Text    string `json:"text"`
            Offset  uint   `json:"offset"`
            Length  uint   `json:"length"`
            Options []struct {
                Text   string         `json:"text"`
                Score  float64        `json:"_score"`
                Source UserSearchBody `json:"_source"`
            } `json:"options"`
        } `json:"suggestLastName,omitempty"`
    } `json:"suggest,omitempty"`
}