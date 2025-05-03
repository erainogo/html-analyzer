package entities

type AnalysisResult struct {
	HTMLVersion  string         `json:"htmlVersion"`
	Title        string         `json:"title"`
	Headings     map[string]int `json:"headings"` // h1-h6
	Links        LinkAnalysis   `json:"links"`
	HasLoginForm bool           `json:"hasLoginForm"`
}

type LinkAnalysis struct {
	Internal     int `json:"internal"`
	External     int `json:"external"`
	Inaccessible int `json:"inaccessible"`
}
