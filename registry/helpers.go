package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func parseFilters(ctx echo.Context) (query.StructuredFilter, error) {
	filterParam := ctx.QueryParam("filter")
	if filterParam == "" {
		return query.StructuredFilter{Logic: query.LogicAnd}, nil
	}
	filterDecodedRaw, err := url.QueryUnescape(filterParam)
	if err != nil {
		return query.StructuredFilter{}, fmt.Errorf("unescaping failed: %w", err)
	}
	filterBytes, err := base64.StdEncoding.DecodeString(filterDecodedRaw)
	if err != nil {
		return query.StructuredFilter{}, fmt.Errorf("base64 decoding failed: %w", err)
	}
	var filterRoot query.StructuredFilter
	if err := json.Unmarshal(filterBytes, &filterRoot); err != nil {
		return query.StructuredFilter{}, fmt.Errorf("JSON unmarshalling failed: %w", err)
	}
	if filterRoot.Logic == "" {
		filterRoot.Logic = query.LogicAnd
	}
	return filterRoot, nil
}

func parseSort(ctx echo.Context) ([]query.SortField, error) {
	sortParam := ctx.QueryParam("sort")
	if sortParam == "" {
		return nil, nil
	}
	sortDecodedRaw, err := url.QueryUnescape(sortParam)
	if err != nil {
		return nil, fmt.Errorf("unescaping failed: %w", err)
	}
	sortBytes, err := base64.StdEncoding.DecodeString(sortDecodedRaw)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding failed: %w", err)
	}
	var sortFields []query.SortField
	if err := json.Unmarshal(sortBytes, &sortFields); err != nil {
		return nil, fmt.Errorf("JSON unmarshalling failed: %w", err)
	}
	for i, sf := range sortFields {
		order := strings.ToLower(strings.TrimSpace(string(sf.Order)))
		if order != "asc" && order != "desc" {
			sortFields[i].Order = "asc"
		} else {
			sortFields[i].Order = query.SortOrder(order)
		}
	}
	return sortFields, nil
}

func parsePageSize(ctx echo.Context) (int, error) {
	pageSize, err := strconv.Atoi(ctx.QueryParam("pageSize"))
	if err != nil {
		return 0, fmt.Errorf("invalid pageSize parameter: %w", err)
	}
	return pageSize, nil
}

func parsePageIndex(ctx echo.Context) (int, error) {
	pageIndex, err := strconv.Atoi(ctx.QueryParam("pageIndex"))
	if err != nil {
		return 0, fmt.Errorf("invalid pageIndex parameter: %w", err)
	}
	return pageIndex, nil
}

func parseQuery(ctx echo.Context) (query.StructuredFilter, int, int, error) {
	filterRoot, err := parseFilters(ctx)
	if err != nil {
		return query.StructuredFilter{}, 0, 0, fmt.Errorf("filter processing failed: %w", err)
	}
	sortFields, err := parseSort(ctx)
	if err != nil {
		return query.StructuredFilter{}, 0, 0, fmt.Errorf("sort processing failed: %w", err)
	}
	filterRoot.SortFields = sortFields
	pageIndex, err := parsePageIndex(ctx)
	if err != nil {
		return query.StructuredFilter{}, 0, 0, fmt.Errorf("pageIndex processing failed: %w", err)
	}
	pageSize, err := parsePageSize(ctx)
	if err != nil {
		return query.StructuredFilter{}, 0, 0, fmt.Errorf("pageSize processing failed: %w", err)
	}

	return filterRoot, pageIndex, pageSize, nil
}

func parseStringQuery(queryStr string) (query.StructuredFilter, int, int, error) {
	values, err := url.ParseQuery(queryStr)
	if err != nil {
		return query.StructuredFilter{}, 0, 0, fmt.Errorf("failed to parse query string: %w", err)
	}
	var filterRoot query.StructuredFilter
	var pageIndex, pageSize int

	if filterParam := values.Get("filter"); filterParam != "" {
		filterDecodedRaw, err := url.QueryUnescape(filterParam)
		if err != nil {
			return query.StructuredFilter{}, 0, 0, fmt.Errorf("unescaping filter failed: %w", err)
		}
		filterBytes, err := base64.StdEncoding.DecodeString(filterDecodedRaw)
		if err != nil {
			return query.StructuredFilter{}, 0, 0, fmt.Errorf("base64 decoding filter failed: %w", err)
		}
		if err := json.Unmarshal(filterBytes, &filterRoot); err != nil {
			return query.StructuredFilter{}, 0, 0, fmt.Errorf("JSON unmarshalling filter failed: %w", err)
		}
	}
	if filterRoot.Logic == "" {
		filterRoot.Logic = query.LogicAnd
	}
	if sortParam := values.Get("sort"); sortParam != "" {
		sortDecodedRaw, err := url.QueryUnescape(sortParam)
		if err != nil {
			return query.StructuredFilter{}, 0, 0, fmt.Errorf("unescaping sort failed: %w", err)
		}
		sortBytes, err := base64.StdEncoding.DecodeString(sortDecodedRaw)
		if err != nil {
			return query.StructuredFilter{}, 0, 0, fmt.Errorf("base64 decoding sort failed: %w", err)
		}
		var sortFields []query.SortField
		if err := json.Unmarshal(sortBytes, &sortFields); err != nil {
			return query.StructuredFilter{}, 0, 0, fmt.Errorf("JSON unmarshalling sort failed: %w", err)
		}
		for i, sf := range sortFields {
			order := strings.ToUpper(string(sf.Order))
			if order != "ASC" && order != "DESC" {
				sortFields[i].Order = "ASC"
			} else {
				sortFields[i].Order = query.SortOrder(order)
			}
		}
		filterRoot.SortFields = sortFields
	}
	if pageIndexParam := values.Get("pageIndex"); pageIndexParam != "" {
		pageIndex, err = strconv.Atoi(pageIndexParam)
		if err != nil {
			return query.StructuredFilter{}, 0, 0, fmt.Errorf("invalid pageIndex parameter: %w", err)
		}
	}
	if pageSizeParam := values.Get("pageSize"); pageSizeParam != "" {
		pageSize, err = strconv.Atoi(pageSizeParam)
		if err != nil {
			return query.StructuredFilter{}, 0, 0, fmt.Errorf("invalid pageSize parameter: %w", err)
		}
	}

	return filterRoot, pageIndex, pageSize, nil
}

func parseUUIDArrayFromQuery(query string) (uuid.UUIDs, bool) {
	if query == "" {
		return nil, false
	}
	query = strings.TrimSpace(query)
	if strings.HasPrefix(query, "[") && strings.HasSuffix(query, "]") {
		query = strings.Trim(query, "[]")
	}
	values := strings.Split(query, ",")
	if len(values) == 0 {
		return nil, false
	}
	var uuids uuid.UUIDs
	for _, value := range values {
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		parsedUUID, err := uuid.Parse(value)
		if err != nil {
			return nil, false
		}
		uuids = append(uuids, parsedUUID)
	}
	return uuids, true
}
