package dropbox

const (
	// ErrTagOther tag name.
	ErrTagOther = "other"
	// ErrTagInternalError tag name.
	ErrTagInternalError = "internal_error"
	// ErrTagNotFound tag name.
	ErrTagNotFound = "not_found"
	// ErrTagNoPermission tag name.
	ErrTagNoPermission = "no_permission"
	// ErrTagNotFile tag name.
	ErrTagNotFile = "not_file"
	// ErrTagNotFolder tag name.
	ErrTagNotFolder = "not_folder"
	// ErrTagRestrictedContent tag name.
	ErrTagRestrictedContent = "restricted_content"
	// ErrTagUnsupportedContentType tag name.
	ErrTagUnsupportedContentType = "unsupported_content_type"
	// ErrTagFile tag name.
	ErrTagFile = "file"
	// ErrTagFolder tag name.
	ErrTagFolder = "folder"
	// ErrTagFileAncestor tag name.
	ErrTagFileAncestor = "file_ancestor"
	// ErrTagNoWritePermission tag name.
	ErrTagNoWritePermission = "no_write_permission"
	// ErrTagInsufficientSpace tag name.
	ErrTagInsufficientSpace = "insufficient_space"
	// ErrTagDisallowedName tag name.
	ErrTagDisallowedName = "disallowed_name"
	// ErrTagTeamFolder tag name.
	ErrTagTeamFolder = "team_folder"
	// ErrTagTooManyWriteOperations tag name.
	ErrTagTooManyWriteOperations = "too_many_write_operations"
	// ErrTagTooManyFiles tag name.
	ErrTagTooManyFiles = "too_many_files"
)

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
	return e.Summary
}

// IsNotFound if error NotFound returns true.
func (e *Error) IsNotFound() bool {
	return (e.Tag == ErrTagNotFound)
}

// IsNoPermission if error NoPermission returns true.
func (e *Error) IsNoPermission() bool {
	return (e.Tag == ErrTagNoPermission)
}
