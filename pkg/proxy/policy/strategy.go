package policy

import (
	"context"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/mitchellh/mapstructure"
	accounts "github.com/owncloud/ocis-accounts/pkg/proto/v0"
)

// Strategy returns a "routing-policy" for a given user-id
type Strategy interface {
	Policy(ctx context.Context, userID string) Name
}

func NewStrategy(cfg PolicyStrategy) Strategy {
	if cfg.Name == "migration" {
		return Migration()
	}

	if cfg.Name == "static_policy" {
		sp := &StaticPolicy{}
		mapstructure.Decode(cfg.Config, sp)
		return &staticPolicy{name: sp.PolicyName}
	}

	return nil
}

// Migration strategy queries the account-service and routes the user to ocis/reva if found.
// If the user is not found it is assumed that he is not migrated and thus still on OC10
func Migration() Strategy {
	return &migrationStrategy{
		accounts.NewSettingsService("com.owncloud.accounts", grpc.NewClient()),
	}
}

type migrationStrategy struct {
	settings accounts.SettingsService
}

// Policy returns "reva" if userId is found in accounts-service, else returns "oc10"
func (ms *migrationStrategy) Policy(ctx context.Context, userID string) Name {
	_, err := ms.settings.Get(ctx, &accounts.Query{Key: userID})
	if err != nil {
		return "oc10"
	}

	return "reva"
}

type staticPolicy struct {
	name Name
}

func (ms *staticPolicy) Policy(ctx context.Context, userID string) Name {
	return ms.name
}
