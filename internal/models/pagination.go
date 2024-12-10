package models

import (
	"errors"
	"net/http"
	"strings"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
)

var (
	ErrInvalidDateFormat = errors.New("invalid date format passed")
)

const (
	LimitParam  = "limit"
	OffsetParam = "offset"
	SortParam   = "sort"
	TagsParam   = "tags"
	SearchParam = "search"
	SinceParam  = "since"
	UntilParam  = "until"
)

type PaginateFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (feed *PaginateFeedQuery) Parse(r *http.Request) error {
	query := r.URL.Query()

	limit := query.Get(LimitParam)
	if limit != "" {
		limitNum, err := httpio.ParseAsInt(limit)
		if err != nil {
			return err
		}
		feed.Limit = limitNum
	}

	offset := query.Get(OffsetParam)
	if offset != "" {
		offsetNum, err := httpio.ParseAsInt(offset)
		if err != nil {
			return err
		}
		feed.Offset = offsetNum
	}

	sort := query.Get(SortParam)
	if sort != "" {
		feed.Sort = sort
	}

	tags := query.Get(TagsParam)
	if tags != "" {
		feed.Tags = strings.Split(tags, ",")
	}

	search := query.Get(SearchParam)
	if search != "" {
		feed.Search = search
	}

	since := query.Get(SinceParam)
	if since != "" {
		sinceParsed, err := httpio.ParseTime(since)
		if err != nil {
			return ErrInvalidDateFormat
		}
		feed.Since = sinceParsed
	}

	until := query.Get(UntilParam)
	if until != "" {
		untilParsed, err := httpio.ParseTime(until)
		if err != nil {
			return ErrInvalidDateFormat
		}
		feed.Until = untilParsed
	}

	return nil
}
