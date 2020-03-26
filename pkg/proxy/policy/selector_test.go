package policy

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/micro/go-micro/v2/client"
	"github.com/owncloud/ocis-accounts/pkg/proto/v0"
	"github.com/owncloud/ocis-pkg/v2/oidc"
	"github.com/owncloud/ocis-proxy/pkg/config"
	"net/http/httptest"
	"testing"
)

func TestStaticSelector(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest("GET", "https://example.org/foo", nil)
	sel := NewStaticSelector(&config.StaticSelectorConf{Policy: "reva"})

	want := "reva"
	got, err := sel(ctx, req)
	if got != want {
		t.Errorf("Expected policy %v got %v", want, got)
	}

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	sel = NewStaticSelector(&config.StaticSelectorConf{Policy: "foo"})

	want = "foo"
	got, err = sel(ctx, req)
	if got != want {
		t.Errorf("Expected policy %v got %v", want, got)
	}

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
}

type testCase struct {
	AccSvcShouldReturnError bool
	Claims                  *oidc.StandardClaims
	Expected                string
}

func TestMigrationSelector(t *testing.T) {
	cfg := config.MigrationSelectorConf{
		AccFoundPolicy:        "found",
		AccNotFoundPolicy:     "not_found",
		UnauthenticatedPolicy: "unauth",
	}
	var tests = []testCase{
		{true, &oidc.StandardClaims{PreferredUsername: "Hans"}, "not_found"},
		{false, &oidc.StandardClaims{PreferredUsername: "Hans"}, "found"},
		{false, nil, "unauth"},
	}

	for k, tc := range tests {
		t.Run(fmt.Sprintf("#%v", k), func(t *testing.T) {
			t.Parallel()
			tc := tc
			sut := NewMigrationSelector(&cfg, mockAccSvc(tc.AccSvcShouldReturnError))
			r := httptest.NewRequest("GET", "https://example.com", nil)
			ctx := oidc.NewContext(r.Context(), tc.Claims)
			nr := r.WithContext(ctx)

			got, err := sut(ctx, nr)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if got != tc.Expected {
				t.Errorf("Expected Policy %v got %v", tc.Expected, got)
			}
		})
	}
}

func mockAccSvc(retErr bool) proto.SettingsService {
	if retErr {
		return &mockSettingsService{
			getFunc: func(ctx context.Context, in *proto.Query, opts ...client.CallOption) (record *proto.Record, err error) {
				return nil, fmt.Errorf("error returned by mockSettingService GET")
			},
		}
	}

	return &mockSettingsService{
		getFunc: func(ctx context.Context, in *proto.Query, opts ...client.CallOption) (record *proto.Record, err error) {
			return &proto.Record{}, nil
		},
	}

}

type mockSettingsService struct {
	setFunc  func(ctx context.Context, in *proto.Record, opts ...client.CallOption) (*proto.Record, error)
	getFunc  func(ctx context.Context, in *proto.Query, opts ...client.CallOption) (*proto.Record, error)
	listFunc func(ctx context.Context, in *empty.Empty, opts ...client.CallOption) (*proto.Records, error)
}

func (m mockSettingsService) Set(ctx context.Context, in *proto.Record, opts ...client.CallOption) (*proto.Record, error) {
	if m.setFunc != nil {
		return m.setFunc(ctx, in, opts...)
	}

	panic("setFunc was called in test but not mocked")
}

func (m mockSettingsService) Get(ctx context.Context, in *proto.Query, opts ...client.CallOption) (*proto.Record, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, in, opts...)
	}

	panic("getFunc was called in test but not mocked")
}

func (m mockSettingsService) List(ctx context.Context, in *empty.Empty, opts ...client.CallOption) (*proto.Records, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, in, opts...)
	}

	panic("listFunc was called in test but not mocked")
}