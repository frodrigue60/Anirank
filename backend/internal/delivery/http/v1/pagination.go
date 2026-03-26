package v1

import (
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// paginatedResponse builds a standardized paginated JSON envelope
// that the SvelteKit frontend expects: { data, pagination: { total, per_page, current_page, last_page, has_more }, links: { self, next, prev } }
func paginatedResponse(c *fiber.Ctx, items interface{}, total int, page, perPage int) fiber.Map {
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	path := c.Path()
	queryParams := c.Queries()
	
	// Strip /api prefix if present to avoid doubling with frontend baseURL
	cleanPath := strings.TrimPrefix(path, "/api")

	buildURL := func(p int) string {
		q := make(map[string]string)
		for k, v := range queryParams {
			q[k] = v
		}
		q["page"] = strconv.Itoa(p)
		
		u := cleanPath + "?"
		for k, v := range q {
			u += k + "=" + v + "&"
		}
		return u[:len(u)-1]
	}

	response := fiber.Map{
		"data": items,
		"pagination": fiber.Map{
			"total":        total,
			"per_page":     perPage,
			"current_page": page,
			"last_page":    lastPage,
			"has_more":     page < lastPage,
		},
		"links": fiber.Map{
			"self": buildURL(page),
		},
	}

	if page < lastPage {
		response["links"].(fiber.Map)["next"] = buildURL(page + 1)
	} else {
		response["links"].(fiber.Map)["next"] = nil
	}

	if page > 1 {
		response["links"].(fiber.Map)["prev"] = buildURL(page - 1)
	} else {
		response["links"].(fiber.Map)["prev"] = nil
	}

	return response
}

func parsePagination(c *fiber.Ctx, defaultLimit int) (int, int) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := defaultLimit
	offset := (page - 1) * limit
	return limit, offset
}
