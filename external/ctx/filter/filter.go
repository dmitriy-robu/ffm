package filter

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Filter struct {
	Search string `json:"search"`
	Page   int64  `json:"page"`
	Limit  int64  `json:"limit"`

	Fields map[string]string `json:"filters"`
}

func GetContextWithFilters(r *http.Request) context.Context {
	var filter Filter

	filter.Search = r.URL.Query().Get("search")

	if page := r.URL.Query().Get("page"); page != "" {
		p, _ := strconv.Atoi(page)
		filter.Page = int64(p)
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		l, _ := strconv.Atoi(limit)
		filter.Limit = int64(l)
	}

	filter.Fields = make(map[string]string)

	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "filters[") && strings.HasSuffix(key, "]") {
			nestedKey := strings.TrimPrefix(strings.TrimSuffix(key, "]"), "filters[")
			if len(values) > 0 {
				filter.Fields[nestedKey] = values[0]
			}
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ctx = context.WithValue(ctx, "filter", filter)

	return ctx
}
