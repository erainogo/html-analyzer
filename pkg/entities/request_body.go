package entities

type RequestBody struct {
	URL         string `json:"url"`
	HTMLContent string `json:"htmlContent,omitempty"`
}
