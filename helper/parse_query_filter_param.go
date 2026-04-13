package common

import (
	"cashier-api/helper/query"
	"strings"
)

func ParseQueryFilterParam(paramSorts string) []*query.QueryFilter {
	var queryFilters []*query.QueryFilter
	if paramSorts != "" {
		sorts := strings.Split(paramSorts, ",")

		for _, s := range sorts {
			parts := strings.Split(s, ":")
			if len(parts) != 2 {
				continue
			}

			queryFilters = append(queryFilters, &query.QueryFilter{
				Column:    parts[0],
				Ascending: parts[1] == "asc",
			})
		}
	}

	return queryFilters
}
