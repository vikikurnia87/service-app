package structs

// Meta contains pagination metadata for API responses.
type Meta struct {
	Count        int  `json:"count"`
	Total        int  `json:"total"`
	TotalPages   int  `json:"total_pages"`
	PerPage      int  `json:"per_page"`
	PreviousPage *int `json:"previous_page"`
	CurrentPage  int  `json:"current_page"`
	NextPage     *int `json:"next_page"`
}

// Pagination holds parsed pagination parameters from query string.
type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

// OrderConfig stores a single sort configuration.
type OrderConfig struct {
	Column    string // database column name
	Direction string // ASC or DESC
}

// OrderMapping maps API field names to database column names.
type OrderMapping map[string]string

// ListParams aggregates pagination, ordering, and search parameters.
type ListParams struct {
	Pagination Pagination
	Orders     []OrderConfig
	Search     string
}
