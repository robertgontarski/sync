package fs

import "strings"

type PathInfo struct {
	IsRemote bool
	User     string
	Host     string
	Path     string
}

// ParsePath parses a path string in the format [user@]host:/path or a local path.
func ParsePath(raw string) PathInfo {
	// Look for the colon separator that indicates a remote path.
	// We need host:/path pattern — the colon must not be part of a Windows drive letter (e.g., C:\).
	colonIdx := strings.Index(raw, ":")
	if colonIdx < 0 {
		return PathInfo{Path: raw}
	}

	// A single letter before colon could be a Windows drive letter (C:\path).
	// Remote hosts are always longer than one character.
	hostPart := raw[:colonIdx]
	if len(hostPart) <= 1 {
		return PathInfo{Path: raw}
	}

	remotePath := raw[colonIdx+1:]

	// Check for user@host
	if atIdx := strings.Index(hostPart, "@"); atIdx >= 0 {
		return PathInfo{
			IsRemote: true,
			User:     hostPart[:atIdx],
			Host:     hostPart[atIdx+1:],
			Path:     remotePath,
		}
	}

	return PathInfo{
		IsRemote: true,
		Host:     hostPart,
		Path:     remotePath,
	}
}
