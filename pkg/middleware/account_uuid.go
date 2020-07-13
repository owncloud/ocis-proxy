package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cs3org/reva/pkg/token/manager/jwt"
	ptoken "github.com/owncloud/ocis-proxy/pkg/middleware/token"

	"github.com/micro/go-micro/v2/client"
	acc "github.com/owncloud/ocis-accounts/pkg/proto/v0"
	"github.com/owncloud/ocis-pkg/v2/log"
	oidc "github.com/owncloud/ocis-pkg/v2/oidc"
	settingsproto "github.com/owncloud/ocis-settings/pkg/proto/v0"
)

func getAccount(l log.Logger, claims *oidc.StandardClaims, ac acc.AccountsService) (account *acc.Account, status int) {
	entry, err := svcCache.Get(AccountsKey, claims.Email)
	if err != nil {
		l.Debug().Msgf("No cache entry for %v", claims.Email)
		resp, err := ac.ListAccounts(context.Background(), &acc.ListAccountsRequest{
			Query:    fmt.Sprintf("mail eq '%s'", strings.ReplaceAll(claims.Email, "'", "''")),
			PageSize: 2,
		})

		if err != nil {
			l.Error().Err(err).Str("email", claims.Email).Msgf("Error fetching from accounts-service")
			status = http.StatusInternalServerError
			return
		}

		if len(resp.Accounts) <= 0 {
			l.Error().Str("email", claims.Email).Msgf("Account not found")
			status = http.StatusNotFound
			return
		}

		// TODO provision account

		if len(resp.Accounts) > 1 {
			l.Error().Str("email", claims.Email).Msgf("More than one account with this email found. Not logging user in.")
			status = http.StatusForbidden
			return
		}

		err = svcCache.Set(AccountsKey, claims.Email, *resp.Accounts[0])
		if err != nil {
			l.Err(err).Str("email", claims.Email).Msgf("Could not cache user")
			status = http.StatusInternalServerError
			return
		}

		account = resp.Accounts[0]
	} else {
		a, ok := entry.V.(acc.Account) // TODO how can we directly point to the cached account?
		if !ok {
			status = http.StatusInternalServerError
			return
		}
		account = &a
	}
	return
}

func createAccount(l log.Logger, claims *oidc.StandardClaims, ac acc.AccountsService) (*acc.Account, int) {
	// TODO check if fields are missing.
	req := &acc.CreateAccountRequest{
		Account: &acc.Account{
			DisplayName:              claims.DisplayName,
			PreferredName:            claims.PreferredUsername,
			OnPremisesSamAccountName: claims.PreferredUsername,
			Mail:                     claims.Email,
			CreationType:             "LocalAccount",
		},
	}
	created, err := ac.CreateAccount(context.Background(), req)
	if err != nil {
		l.Error().Err(err).Interface("account", req.Account).Msg("could not create account")
		return nil, http.StatusInternalServerError
	}

	return created, 0
}

func getRoles(ctx context.Context, accID string) (*settingsproto.UserRoleAssignments, error) {
	rs := settingsproto.NewRoleService("com.owncloud.api.settings", client.DefaultClient)
	roles, err := rs.ListRoleAssignments(ctx, &settingsproto.ListRoleAssignmentsRequest{
		Assignment: &settingsproto.RoleAssignmentIdentifier{
			AccountUuid: accID,
		},
	})
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// AccountUUID provides a middleware which mints a jwt and adds it to the proxied request based
// on the oidc-claims
func AccountUUID(opts ...Option) func(next http.Handler) http.Handler {
	opt := newOptions(opts...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := opt.Logger
			claims := oidc.FromContext(r.Context())
			if claims == nil {
				next.ServeHTTP(w, r)
				return
			}

			// TODO allow lookup by username?
			// TODO allow lookup by custom claim, eg an id

			account, status := getAccount(l, claims, opt.AccountsClient)
			if status != 0 {
				if status == http.StatusNotFound {
					account, status = createAccount(l, claims, opt.AccountsClient)
					if status != 0 {
						w.WriteHeader(status)
						return
					}
				} else {
					w.WriteHeader(status)
					return
				}
			}
			if !account.AccountEnabled {
				l.Debug().Interface("account", account).Msg("account is disabled")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			groups := make([]string, len(account.MemberOf))
			for i := range account.MemberOf {
				// reva needs the unix group name
				groups[i] = account.MemberOf[i].OnPremisesSamAccountName
			}

			roles, err := getRoles(r.Context(), account.Id)
			if err != nil {
				l.Debug().Interface("account", account).Msg("account is disabled")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// TODO: this requires an explanation. A PR against Reva is pending to check whether
			// or not we opt for this implementation. This adds the roles to the context serialized
			// as []string and reva will check if it is present, if so, will add `roles` to the list of claims.
			ctx := contextWithRoles(r.Context(), roles)
			l.Debug().Interface("claims", claims).Interface("account", account).Msgf("Associated claims with uuid")
			token, err := mintToken(ctx, account, groups, opt)
			if err != nil {
				l.Debug().Str("minting", "account_uuid Middleware").Msgf("%v", err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			l.Debug().Str("access-token", token).Msg("access token contents")
			r.Header.Set("x-access-token", token)
			next.ServeHTTP(w, r)
		})
	}
}

func contextWithRoles(c context.Context, roles *settingsproto.UserRoleAssignments) context.Context {
	var r []string
	for _, assignment := range roles.Assignments {
		r = append(r, assignment.Role)
	}
	return context.WithValue(c, jwt.RolesKey{}, r)
}

func mintToken(ctx context.Context, account *acc.Account, groups []string, opts Options) (string, error) {
	i := ptoken.Info{
		Account: account,
		Groups:  groups,
	}
	token, err := i.MintRevaFromInfo(ctx, opts.TokenManagerConfig.JWTSecret)

	if err != nil {
		return "", fmt.Errorf("could not mint token: %v", err.Error())
	}

	return token, nil
}
