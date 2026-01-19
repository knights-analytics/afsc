package abs

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if responseError, ok := err.(*azcore.ResponseError); ok {
		if responseError.ErrorCode == "BlobNotFound" || responseError.ErrorCode == "ContainerNotFound" || responseError.StatusCode == 404 {
			return true
		}
	}
	return strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "404")
}
