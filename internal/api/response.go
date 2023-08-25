package api

import (
	"fmt"
	"io"
	"net/http"
)

func ExtractErrorDetails(resp *http.Response, err error) error {
	if err == nil {
		return nil
	}

	if resp == nil {
		return err
	}

	rawbody, parseErr := io.ReadAll(resp.Body)
	if parseErr != nil {
		return fmt.Errorf("%v (failed parsing response body: %v)", err, parseErr)
	}
	return fmt.Errorf("%v (%s)", err, string(rawbody))
}
