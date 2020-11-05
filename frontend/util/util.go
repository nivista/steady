package util

import (
	"context"
	"encoding/base64"
	"path"
	"strings"
)

type key int

const (
	clientIDKey key = iota
)

// SetClientID returns a new context with clientID set.
func SetClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDKey, clientID)
}

// GetClientID returns the clientID and true if the context has that clientID, otherwise ok is false.
func GetClientID(ctx context.Context) (clientID string, ok bool) {
	v, ok := ctx.Value(clientIDKey).(string)
	return v, ok
}

// ParseBasicAuth parses an HTTP Basic Authentication string.
// copied from "net/http"
func ParseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

// ShiftPath returns the topmost directory of the path, and the rest of the path.
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
