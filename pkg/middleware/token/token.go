package token

import (
	"context"
	"fmt"

	revauser "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	"github.com/cs3org/reva/pkg/token/manager/jwt"
	acc "github.com/owncloud/ocis-accounts/pkg/proto/v0"
)

// RolesKey stores roles on a context.
// TODO(refs): should be moved onto reva, consumers should depend on the library,
// not the other way around.
type RolesKey struct{}

// Info encapsulates ocis accounts related data.
type Info struct {
	Account *acc.Account
	Groups  []string
}

// MintRevaFromInfo uses reva's JWT manager to mint a token.
// roles need to be written to the request context in order for Reva to add them to the token.
func (i Info) MintRevaFromInfo(ctx context.Context, s string) (string, error) {
	tokenManager, err := jwt.New(map[string]interface{}{
		"secret":  s,
		"expires": int64(60),
	})
	if err != nil {
		return "", fmt.Errorf("Could not initialize token-manager")
	}

	token, err := tokenManager.MintToken(ctx, &revauser.User{
		Id: &revauser.UserId{
			OpaqueId: i.Account.Id,
		},
		Username:     i.Account.OnPremisesSamAccountName,
		DisplayName:  i.Account.DisplayName,
		Mail:         i.Account.Mail,
		MailVerified: i.Account.ExternalUserState == "" || i.Account.ExternalUserState == "Accepted",
		Groups:       i.Groups,
	})

	return token, nil
}
