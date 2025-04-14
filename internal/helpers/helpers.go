package helpers

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func ReadJSON(c echo.Context, dst any) error {
	// Set maximum bytes limit
	maxBytes := 1_048_576
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, int64(maxBytes))

	// Create JSON decoder with strict settings
	dec := json.NewDecoder(c.Request().Body)
	dec.DisallowUnknownFields()

	// Attempt to decode
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("body contains badly formed JSON (at character %d)", syntaxError.Offset))
		case errors.Is(err, io.ErrUnexpectedEOF):
			return echo.NewHTTPError(http.StatusBadRequest, "body contains badly formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field))
			}
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset))
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("body contains unknown key %s", fieldName))
		case errors.Is(err, io.EOF):
			return echo.NewHTTPError(http.StatusBadRequest, "body must not be empty")
		case errors.As(err, &maxBytesError):
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge,
				fmt.Sprintf("body must not be larger than %d bytes", maxBytes))
		case errors.As(err, &invalidUnmarshalError):
			// This should never happen and indicates a programming error
			panic(err)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	// Check for multiple JSON values
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return echo.NewHTTPError(http.StatusBadRequest, "body must only contain a single JSON value")
	}

	return nil
}

func Md5(data string) []byte {
	return []byte(data)
	h := md5.New()
	h.Write([]byte(data))
	return h.Sum(nil)
	//return hex.EncodeToString(h.Sum(nil))
}
