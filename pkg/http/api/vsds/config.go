package vsds

import (
	"github.com/gczuczy/ed-survey-tools/pkg/gcp"
	"github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

// VSDSConfigResponse is the response type for GET /api/vsds/config
type VSDSConfigResponse struct {
	GCPClientEmail string `json:"gcp_client_email"`
}

func getConfig(r *wrappers.Request) wrappers.IResponse {
	return wrappers.Success(VSDSConfigResponse{
		GCPClientEmail: gcp.ClientEmail(),
	})
}
