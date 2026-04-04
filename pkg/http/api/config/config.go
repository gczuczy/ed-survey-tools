package config

import (
	"github.com/gczuczy/ed-survey-tools/pkg/config"
	w "github.com/gczuczy/ed-survey-tools/pkg/http/wrappers"
)

var bundlesCfg *config.BundlesConfig

// Init stores the BundlesConfig for use by handlers in this package.
func Init(cfg *config.BundlesConfig) {
	bundlesCfg = cfg
}

// AppConfigResponse is the response type for GET /api/config
type AppConfigResponse struct {
	BundleBaseURL string `json:"bundleBaseUrl"`
}

func getConfig(r *w.Request) w.IResponse {
	return w.Success(AppConfigResponse{
		BundleBaseURL: bundlesCfg.BaseURL,
	})
}
