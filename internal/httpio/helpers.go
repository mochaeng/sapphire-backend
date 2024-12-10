package httpio

import (
	"strconv"
	"time"
)

func ParseTime(date string) (string, error) {
	parsedTime, err := time.Parse(time.DateTime, date)
	if err != nil {
		return "", ErrInvalidSearchParamType
	}
	return parsedTime.Format(time.DateTime), nil
}

func ParseAsInt(param string) (int, error) {
	paramNum, err := strconv.Atoi(param)
	if err != nil {
		return 0, ErrInvalidSearchParamType
	}
	return paramNum, nil
}
