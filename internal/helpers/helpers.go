package helpers

import (
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"service-app/internal/structs"
)

// GetPaginationParams parses limit and page from the query string.
// If values are missing or invalid, defaults are used (page=1, limit=defaultLimit).
func GetPaginationParams(c *echo.Context, defaultLimit int) structs.Pagination {
	perPageParam := c.QueryParam("limit")
	pageParam := c.QueryParam("page")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	limit := defaultLimit
	if perPageParam != "" {
		limit, err = strconv.Atoi(perPageParam)
		if err != nil || limit < 1 {
			limit = defaultLimit
		}
	}

	offset := (page - 1) * limit

	return structs.Pagination{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

// BuildPaginationMeta constructs a Meta struct from pagination data.
func BuildPaginationMeta(count, total, currentPage, perPage int) structs.Meta {
	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	var prevPage *int
	if currentPage > 1 {
		p := currentPage - 1
		prevPage = &p
	}

	var nextPage *int
	if currentPage < totalPages {
		n := currentPage + 1
		nextPage = &n
	}

	return structs.Meta{
		Count:        count,
		Total:        total,
		TotalPages:   totalPages,
		PerPage:      perPage,
		PreviousPage: prevPage,
		CurrentPage:  currentPage,
		NextPage:     nextPage,
	}
}

// GetPaginationTTL returns a cache TTL based on the page number.
// Earlier pages are cached longer since they are accessed more frequently.
func GetPaginationTTL(page int) time.Duration {
	switch {
	case page == 1:
		return 10 * time.Minute // Page 1: 10 minutes
	case page <= 5:
		return 5 * time.Minute // Page 2-5: 5 minutes
	case page <= 10:
		return 3 * time.Minute // Page 6-10: 3 minutes
	default:
		return 3 * time.Minute // Page 11+: 3 minutes
	}
}

// ParseOrderParams extracts all order parameters from the query string.
// Format: order[field]=ASC|DESC
// Example: order[created_at]=DESC&order[name]=ASC
// Only fields present in allowedFields are accepted.
func ParseOrderParams(c *echo.Context, allowedFields structs.OrderMapping) []structs.OrderConfig {
	var orders []structs.OrderConfig

	queryParams := c.QueryParams()

	for key, values := range queryParams {
		// Check if key starts with "order[" and ends with "]"
		if !strings.HasPrefix(key, "order[") || !strings.HasSuffix(key, "]") {
			continue
		}

		// Extract field name from order[fieldName]
		fieldName := key[6 : len(key)-1] // Remove "order[" and "]"

		if fieldName == "" || len(values) == 0 {
			continue
		}

		// Check if the field is allowed
		dbColumn, allowed := allowedFields[fieldName]
		if !allowed {
			continue
		}

		// Validate direction
		direction := strings.ToUpper(strings.TrimSpace(values[0]))
		if direction != "ASC" && direction != "DESC" {
			direction = "ASC" // default
		}

		orders = append(orders, structs.OrderConfig{
			Column:    dbColumn,
			Direction: direction,
		})
	}

	return orders
}

// ParseOrderParamsWithDefault works like ParseOrderParams but falls back
// to defaultOrders when no valid order parameters are found.
func ParseOrderParamsWithDefault(c *echo.Context, allowedFields structs.OrderMapping, defaultOrders []structs.OrderConfig) []structs.OrderConfig {
	orders := ParseOrderParams(c, allowedFields)

	if len(orders) == 0 {
		return defaultOrders
	}

	return orders
}
