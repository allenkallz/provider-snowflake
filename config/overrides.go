package config

import ujconfig "github.com/crossplane/upjet/pkg/config"

func GroupOverrides(r *ujconfig.Resource) {

	apiGroup, ok := resourceApiGroupConfig[r.Name]

	if ok {
		r.ShortGroup = apiGroup
	}

}
