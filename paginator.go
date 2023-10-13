package paginator

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

type Cursor map[string]interface{}

type Pagination struct {
	NextCursor string `json:"nextCursor"`
	PrevCursor string `json:"prevCursor"`
}

func NewCursor(id int64, createdAt time.Time, pointsNext bool) Cursor {
	return Cursor{
		"id":         id,
		"createdAt":  createdAt,
		"pointsNext": pointsNext,
	}
}

func Reverse[T any](s []T) []T {
	for i := 0; i < len(s)/2; i++ {
		s[i], s[len(s)-1-i] = s[len(s)-1-i], s[i]
	}
	return s
}

func generatePager(next, prev Cursor) *Pagination {
	return &Pagination{
		NextCursor: encodeCursor(next),
		PrevCursor: encodeCursor(prev),
	}
}

func encodeCursor(cursor Cursor) string {
	if len(cursor) == 0 {
		return ""
	}
	serializedCursor, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	encodedCursor := base64.StdEncoding.EncodeToString(serializedCursor)
	return encodedCursor
}

func DecodeCursor(cursor string) (Cursor, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	var cur Cursor
	if err := json.Unmarshal(decodedCursor, &cur); err != nil {
		return nil, err
	}
	return cur, nil
}

func GetPaginationOperator(pointsNext bool, sortOrder string) (string, string) {
	sortOrder = strings.ToLower(sortOrder)
	if (pointsNext && sortOrder == "asc") || (!pointsNext && sortOrder == "desc") {
		return ">", "asc"
	}
	if (pointsNext && sortOrder == "desc") || (!pointsNext && sortOrder == "asc") {
		return "<", "desc"
	}

	return "", ""
}

type Items struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `pg:"created_at"`
}

func CalculatePagination(isFirstPage bool, limit int, items []*Items, isLastPage bool) *Pagination {
	if limit == 0 {
		return nil
	}
	if isFirstPage && isLastPage {
		return nil
	}

	if hasPagination := len(items)+1 > limit; !hasPagination {
		isLastPage = true
		num := len(items)
		if num == 0 {
			return nil
		}

		lastItem := items[num-1]
		return generatePager(nil, NewCursor(lastItem.ID, lastItem.CreatedAt, false))
	}

	lastItem := items[limit-1]
	firstItem := items[0]
	nextCur := NewCursor(lastItem.ID, lastItem.CreatedAt, true)
	prevCur := NewCursor(firstItem.ID, firstItem.CreatedAt, false)

	if isFirstPage {
		return generatePager(nextCur, nil)
	}
	if isLastPage {
		return generatePager(nil, prevCur)
	}

	return generatePager(nextCur, prevCur)
}
