package pagination

import (
	"database/sql"
	"errors"
)

var (
	ErrInvalidDateFormat = errors.New("invalid date format passed")
)

const (
	TimelineLimitDefault = 10
	TimeLineLimitMax     = 20
)

type PaginateFeedQuery struct {
	Limit      int
	Cursor     sql.NullTime
	NextCursor string
}

func (feed *PaginateFeedQuery) Parse(limitParam, cursorParam string) error {
	limit, err := parseLimit(limitParam, TimelineLimitDefault, TimeLineLimitMax)
	if err != nil {
		return err
	}

	cursor, err := parseCursor(cursorParam)
	if err != nil {
		return err
	}

	feed.Limit = *limit
	feed.Cursor = *cursor

	return nil
}
