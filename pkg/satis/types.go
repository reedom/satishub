package satis

// PackageInfo contains package information to be managed by satis.
type PackageInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	URL     string `json:"url,omitempty"`
	Type    string `json:"type,omitempty"`
}

// ServiceResult represents a result of Service tasks.
type ServiceResult struct {
	Error error
}

// Succeeded determines whether the service task has succeeded.
func (s ServiceResult) Succeeded() bool {
	return s.Error == nil
}
