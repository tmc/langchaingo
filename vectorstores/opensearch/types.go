package opensearch

type searchResults struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64            `json:"max_score"`
		Hits     []searchResultsHit `json:"hits"`
	} `json:"hits"`
}

type searchResultsHit struct {
	Index  string   `json:"_index"`
	ID     string   `json:"_id"`
	Score  float32  `json:"_score"`
	Source document `json:"_source"`
}
