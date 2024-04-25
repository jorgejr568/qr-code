package validators

import "net/url"

func ValidateUrl(v string) bool {
	_, err := url.ParseRequestURI(v)
	return err == nil
}
