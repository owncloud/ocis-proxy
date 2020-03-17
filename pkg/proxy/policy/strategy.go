package policy

import (
	"context"
	"fmt"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/mitchellh/mapstructure"
	accounts "github.com/owncloud/ocis-accounts/pkg/proto/v0"
)

// Strategy returns a "routing-policy" for a given user-id
type Strategy interface {
	Policy(ctx context.Context, userID string) Name
}

// NewStrategy creates a policy-strategy from a given weakly-typed configuration.
func NewStrategy(cfg StrategyConfig) (Strategy, error) {
	if cfg.Name == "migration" {
		return MigrationStrategy(), nil
	}

	if cfg.Name == "static_policy" {
		sp := &StaticPolicyConfig{}
		err := mapstructure.Decode(cfg.Config, sp)
		if err != nil {
			return nil, err

		}
		return StaticPolicyStrategy(sp), nil
	}

	return nil, fmt.Errorf("invalid policy strategy type %v", cfg.Name)
}

// MigrationStrategy strategy creates a strategy which queries the account-service and routes the user to ocis/reva if found.
// If the user is not found it is assumed that he is not migrated and thus still on OC10.
func MigrationStrategy() Strategy {
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

// StaticPolicyStrategy uses a statically configured policy-name
func StaticPolicyStrategy(cfg *StaticPolicyConfig) Strategy {
	return &staticPolicyStrategy{
		name: cfg.PolicyName,
	}
}

type staticPolicyStrategy struct {
	name Name
}

func (ms *staticPolicyStrategy) Policy(ctx context.Context, userID string) Name {
	return ms.name
}
