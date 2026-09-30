package domain

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// EncodeSpecID and EncodeChangeID pack a project registry id with a
// capability path or change name into a single opaque, URL-safe string, so
// the REST API can address a spec/change with one path segment instead of a
// project path (which may contain slashes).

func EncodeSpecID(projectID int64, capability string) string {
	return encodeID(projectID, capability)
}

func DecodeSpecID(id string) (projectID int64, capability string, err error) {
	return decodeID(id)
}

func EncodeChangeID(projectID int64, name string) string {
	return encodeID(projectID, name)
}

func DecodeChangeID(id string) (projectID int64, name string, err error) {
	return decodeID(id)
}

func encodeID(projectID int64, suffix string) string {
	raw := fmt.Sprintf("%d:%s", projectID, suffix)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeID(id string) (int64, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return 0, "", fmt.Errorf("invalid id: %w", err)
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid id: %q", string(raw))
	}
	projectID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid id: %w", err)
	}
	return projectID, parts[1], nil
}
