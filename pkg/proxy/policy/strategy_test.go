package policy

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/micro/go-micro/v2/client"
	"github.com/owncloud/ocis-accounts/pkg/proto/v0"
	"testing"
)

func TestMigrationStrategy(t *testing.T) {
	strategy := &migrationStrategy{&mockSettingsService{
		getFunc: func(ctx context.Context, in *proto.Query, opts ...client.CallOption) (record *proto.Record, err error) {
			return nil, fmt.Errorf("accounts not reachable %w", errors.New("user Not found"))
		},
	}}

	res := strategy.Policy(context.Background(), "foo")
	if res != "oc10" {
		t.Errorf("Call to Policy() should return 'oc10' got %v", res)
	}

	strategy = &migrationStrategy{&mockSettingsService{
		getFunc: func(ctx context.Context, in *proto.Query, opts ...client.CallOption) (record *proto.Record, err error) {
			return nil, nil
		},
	}}

	res = strategy.Policy(context.Background(), "foo")
	if res != "reva" {
		t.Errorf("Call to Policy() should return 'reva' got %v", res)
	}
}

func TestStaticPolicyStrategy(t *testing.T) {
	strategy := StaticPolicyStrategy(&StaticPolicyConfig{PolicyName: "policyA"})

	got := strategy.Policy(context.Background(), "foo")
	if got != "policyA" {
		t.Errorf("Call to Policy() should return 'policyA' got %v", got)
	}

	strategy = StaticPolicyStrategy(&StaticPolicyConfig{PolicyName: "policyB"})

	got = strategy.Policy(context.Background(), "foo")
	if got != "policyB" {
		t.Errorf("Call to Policy() should return 'policyB' got %v", got)
	}

}

type mockSettingsService struct {
	setFunc  func(ctx context.Context, in *proto.Record, opts ...client.CallOption) (*proto.Record, error)
	getFunc  func(ctx context.Context, in *proto.Query, opts ...client.CallOption) (*proto.Record, error)
	listFunc func(ctx context.Context, in *empty.Empty, opts ...client.CallOption) (*proto.Records, error)
}

func (t *mockSettingsService) Set(ctx context.Context, in *proto.Record, opts ...client.CallOption) (*proto.Record, error) {
	if t.setFunc != nil {
		return t.setFunc(ctx, in, opts...)
	}

	panic("setFunc was called in test but not mocked")
}

func (t *mockSettingsService) Get(ctx context.Context, in *proto.Query, opts ...client.CallOption) (*proto.Record, error) {
	if t.getFunc != nil {
		return t.getFunc(ctx, in, opts...)
	}

	panic("getFunc was called in test but not mocked")

}

func (t *mockSettingsService) List(ctx context.Context, in *empty.Empty, opts ...client.CallOption) (*proto.Records, error) {
	if t.listFunc != nil {
		return t.listFunc(ctx, in, opts...)
	}

	panic("listFunc was called in test but not mocked")

}
