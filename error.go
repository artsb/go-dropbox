package dropbox

import "fmt"

// errorInfo Dropbox error info.
type errorInfo struct {
	Summary string `json:"error_summary"`
	Error   struct {
		Tag string `json:".tag"`
	} `json:"error"`
}

// Error response.
type Error struct {
	Status     string
	StatusCode int
	Summary    string
	Tag        string
}

// Error string.
func (e *Error) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("dropbox: Status: %d (%s); Tag: %s; Summary: %s", e.StatusCode, e.Status, e.Tag, e.Summary)
	}
	return fmt.Sprintf("dropbox: Tag: %s; Summary: %s", e.Tag, e.Summary)
}
