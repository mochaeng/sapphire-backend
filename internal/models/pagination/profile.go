package pagination

import (
	"database/sql"
)

const (
	ProfileLimitDefault = 10
	ProfileLimitMax     = 15
)

var DefaultCursor = sql.NullTime{Valid: false}

type UserPosts struct {
	UserID     int64
	Limit      int
	Cursor     sql.NullTime
	Username   string
	FirstName  string
	LastName   string
	NextCursor string
}

func (payload *UserPosts) Parser(limitParam, cursorParam string) error {
	limit, err := parseLimit(limitParam, ProfileLimitDefault, ProfileLimitMax)
	if err != nil {
		return err
	}

	cursor, err := parseCursor(cursorParam)
	if err != nil {
		return err
	}

	payload.Limit = *limit
	payload.Cursor = *cursor

	return nil
}
