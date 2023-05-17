package opts // import "github.com/docker/docker/runconfig/opts"

import (
	"github.com/docker/docker/mystrings"
)

// ConvertKVStringsToMap converts ["key=value"] to {"key":"value"}
func ConvertKVStringsToMap(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		k, v, _ := mystrings.Cut(value, "=")
		result[k] = v
	}

	return result
}
