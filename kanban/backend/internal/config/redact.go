package config

import "strings"

// RedactDatabaseURL masks the password in a postgres connection string so it
// can safely appear in logs.
func RedactDatabaseURL(raw string) string {
	schemeEnd := strings.Index(raw, "://")
	if schemeEnd < 0 {
		return raw
	}
	rest := raw[schemeEnd+3:]

	// Only the authority section (before the first '/') can contain credentials.
	authEnd := strings.Index(rest, "/")
	authority := rest
	tail := ""
	if authEnd >= 0 {
		authority, tail = rest[:authEnd], rest[authEnd:]
	}

	at := strings.LastIndex(authority, "@")
	if at < 0 {
		return raw // no credentials present
	}
	userinfo := authority[:at]
	host := authority[at+1:]

	if colon := strings.Index(userinfo, ":"); colon >= 0 {
		userinfo = userinfo[:colon] + ":****"
	}
	return raw[:schemeEnd+3] + userinfo + "@" + host + tail
}
