package common

import "io"

func ReadBody(body io.ReadCloser) (string, error) {
	bytes, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	result := string(bytes)

	return result, nil
}
