package config

import "derand-cli/profile"

type LocalProfileInfo struct {
	Path string               `json:"path"`
	Data profile.LocalProfile `json:"data"`
}
