package pagination

import (
	"database/sql"
	"time"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
)

const ParametersMaxSize = 50

func parseLimit(limitParam string, limitDefault, limitMax int) (*int, error) {
	limit := limitDefault
	if limitParam != "" {
		if len(limitParam) < ParametersMaxSize && len(limitParam) > 0 {
			numParsed, err := httpio.ParseAsInt(limitParam)
			if err != nil {
				return nil, httpio.ErrInvalidSearchParamType
			}
			if numParsed > limitMax {
				numParsed = limitMax
			}
			limit = numParsed
		}
	}
	return &limit, nil
}

func parseCursor(cursorParam string) (*sql.NullTime, error) {
	cursor := DefaultCursor
	if cursorParam != "" {
		if len(cursorParam) < ParametersMaxSize && len(cursorParam) > 0 {
			parsedTime, err := time.Parse(time.RFC3339Nano, cursorParam)
			if err != nil {
				return nil, httpio.ErrInvalidSearchParamType
			}
			cursor = sql.NullTime{Time: parsedTime, Valid: true}
		}
	}
	return &cursor, nil
}
