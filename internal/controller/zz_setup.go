// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	account "github.com/allenkallz/provider-snowflake/internal/controller/account/account"
	accountrole "github.com/allenkallz/provider-snowflake/internal/controller/account/accountrole"
	database "github.com/allenkallz/provider-snowflake/internal/controller/database/database"
	databaserole "github.com/allenkallz/provider-snowflake/internal/controller/database/databaserole"
	fileformat "github.com/allenkallz/provider-snowflake/internal/controller/database/fileformat"
	pipe "github.com/allenkallz/provider-snowflake/internal/controller/database/pipe"
	stage "github.com/allenkallz/provider-snowflake/internal/controller/database/stage"
	providerconfig "github.com/allenkallz/provider-snowflake/internal/controller/providerconfig"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		account.Setup,
		accountrole.Setup,
		database.Setup,
		databaserole.Setup,
		fileformat.Setup,
		pipe.Setup,
		stage.Setup,
		providerconfig.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
