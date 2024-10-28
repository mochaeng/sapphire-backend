package httpio

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/mochaeng/sapphire-backend/internal/media"
)

func ReadFormDataValue(r *http.Request, data any) error {
	dataValuesPtr := reflect.ValueOf(data)
	if dataValuesPtr.Kind() != reflect.Ptr || dataValuesPtr.IsNil() {
		return ErrWrongParameterType
	}
	dataValues := reflect.Indirect(dataValuesPtr)
	typeOfData := dataValues.Type()
	formData := make(map[string]any, dataValues.NumField())
	for i := 0; i < dataValues.NumField(); i++ {
		fieldName := typeOfData.Field(i).Name
		fieldType := typeOfData.Field(i).Type
		fieldKind := fieldType.Kind()
		key := strings.ToLower(fieldName)
		formValues := r.Form[key]
		if fieldKind == reflect.Slice || fieldKind == reflect.Array {
			formData[key] = noEmptyTags(formValues)
		} else {
			if len(formValues) == 1 {
				formData[key] = formValues[0]
			}
		}
	}
	jsonData, err := json.Marshal(formData)
	if err != nil {
		return ErrMarshalData
	}
	if err := json.Unmarshal(jsonData, data); err != nil {
		return ErrMarshalData
	}
	return nil
}

func ReadFormFiles(r *http.Request, fileField string, maxReadSize int64) ([]byte, error) {
	file, fileHeader, err := r.FormFile(fileField)
	if err != nil {
		return nil, nil
	}
	defer file.Close()
	if fileHeader.Size > maxReadSize {
		return nil, media.ErrFileTooBig
	}
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, media.ErrInvalidFile
	}
	return fileBytes, nil
}

func noEmptyTags(arr []string) []string {
	newArr := []string{}
	for _, value := range arr {
		if len(value) != 0 {
			newArr = append(newArr, value)
		}
	}
	return newArr
}
