//go:build unit

/**
* @license
* Copyright 2026 Dynatrace LLC
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package ipallowlist

import (
	"context"
	"errors"
	"fmt"
	"testing"

	ipallowlist "github.com/dynatrace-oss/terraform-provider-dynatrace/dynatrace/api/env/ipallowlist/settings"
	"github.com/dynatrace-oss/terraform-provider-dynatrace/dynatrace/api/iam"
	"github.com/stretchr/testify/assert"
)

const (
	testAccountID   = "test-account-id"
	testEndpointURL = "https://api-test.dynatrace.com"
)

type mockIAMClient struct {
	GETFunc              func(ctx context.Context, url string, expectedResponseCode int, forceNewBearer bool) ([]byte, error)
	PUTMultiResponseFunc func(ctx context.Context, url string, payload any, expectedResponseCodes []int, forceNewBearer bool) ([]byte, error)
}

func (me *mockIAMClient) POST(context.Context, string, any, int, bool) ([]byte, error) {
	panic("mock does not support POST")
}

func (me *mockIAMClient) PUT(context.Context, string, any, int, bool) ([]byte, error) {
	panic("mock does not support PUT")
}

func (me *mockIAMClient) PUT_MULTI_RESPONSE(ctx context.Context, url string, payload any, expectedResponseCodes []int, forceNewBearer bool) ([]byte, error) {
	return me.PUTMultiResponseFunc(ctx, url, payload, expectedResponseCodes, forceNewBearer)
}

func (me *mockIAMClient) GET(ctx context.Context, url string, expectedResponseCode int, forceNewBearer bool) ([]byte, error) {
	return me.GETFunc(ctx, url, expectedResponseCode, forceNewBearer)
}

func (me *mockIAMClient) DELETE(context.Context, string, int, bool) ([]byte, error) {
	panic("mock does not support DELETE")
}

func (me *mockIAMClient) DELETE_MULTI_RESPONSE(context.Context, string, []int, bool) ([]byte, error) {
	panic("mock does not support DELETE_MULTI_RESPONSE")
}

type mockIAMClientGetter struct {
	client *mockIAMClient
}

func (me *mockIAMClientGetter) New(_ context.Context) iam.IAMClient {
	return me.client
}

func createTestService(client *mockIAMClient) *ServiceClient {
	return &ServiceClient{
		iamClientGetter: &mockIAMClientGetter{client: client},
		accountID:       testAccountID,
		endpointURL:     testEndpointURL,
	}
}

func TestService_SchemaID(t *testing.T) {
	service := createTestService(&mockIAMClient{})
	assert.Equal(t, "accounts:environment:ip-allowlist", service.SchemaID())
}

func TestService_Create(t *testing.T) {
	mockClient := &mockIAMClient{
		PUTMultiResponseFunc: func(ctx context.Context, url string, payload any, expectedResponseCodes []int, forceNewBearer bool) ([]byte, error) {
			assert.Equal(t, fmt.Sprintf("%s/env/v1/accounts/environments/%s/ip-allowlist", testEndpointURL, "abc123"), url)
			assert.ElementsMatch(t, []int{200, 201, 204}, expectedResponseCodes)
			cfg := payload.(*ipallowlist.IPAllowlist)
			assert.True(t, cfg.Enabled)
			assert.True(t, cfg.AllowWebhookOverride)
			assert.Len(t, cfg.Allowlist, 1)
			return nil, nil
		},
	}

	service := createTestService(mockClient)
	cfg := &ipallowlist.IPAllowlist{
		EnvironmentID:        "abc123",
		Enabled:              true,
		AllowWebhookOverride: true,
		Allowlist: []ipallowlist.AllowlistEntry{{
			Name:    "office",
			IPRange: "10.0.0.0/8",
		}},
	}

	stub, err := service.Create(t.Context(), cfg)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", stub.ID)
}

func TestService_Get(t *testing.T) {
	mockClient := &mockIAMClient{
		GETFunc: func(ctx context.Context, url string, expectedResponseCode int, forceNewBearer bool) ([]byte, error) {
			assert.Equal(t, fmt.Sprintf("%s/env/v1/accounts/environments/%s/ip-allowlist", testEndpointURL, "abc123"), url)
			assert.Equal(t, 200, expectedResponseCode)
			return []byte(`{"enabled":true,"allowWebhookOverride":false,"allowlist":[{"name":"office","ipRange":"10.0.0.0/8"}]}`), nil
		},
	}

	service := createTestService(mockClient)
	cfg := &ipallowlist.IPAllowlist{}
	err := service.Get(t.Context(), "abc123", cfg)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", cfg.EnvironmentID)
	assert.True(t, cfg.Enabled)
	assert.False(t, cfg.AllowWebhookOverride)
	assert.Equal(t, "office", cfg.Allowlist[0].Name)
}

func TestService_Update(t *testing.T) {
	mockClient := &mockIAMClient{
		PUTMultiResponseFunc: func(ctx context.Context, url string, payload any, expectedResponseCodes []int, forceNewBearer bool) ([]byte, error) {
			assert.Equal(t, fmt.Sprintf("%s/env/v1/accounts/environments/%s/ip-allowlist", testEndpointURL, "abc123"), url)
			cfg := payload.(*ipallowlist.IPAllowlist)
			assert.True(t, cfg.Enabled)
			return nil, nil
		},
	}

	service := createTestService(mockClient)
	cfg := &ipallowlist.IPAllowlist{EnvironmentID: "abc123", Enabled: true}
	assert.NoError(t, service.Update(t.Context(), "abc123", cfg))
}

func TestService_Delete(t *testing.T) {
	mockClient := &mockIAMClient{
		PUTMultiResponseFunc: func(ctx context.Context, url string, payload any, expectedResponseCodes []int, forceNewBearer bool) ([]byte, error) {
			assert.Equal(t, fmt.Sprintf("%s/env/v1/accounts/environments/%s/ip-allowlist", testEndpointURL, "abc123"), url)
			cfg := payload.(*ipallowlist.IPAllowlist)
			assert.False(t, cfg.Enabled)
			assert.False(t, cfg.AllowWebhookOverride)
			assert.Empty(t, cfg.Allowlist)
			return nil, nil
		},
	}

	service := createTestService(mockClient)
	assert.NoError(t, service.Delete(t.Context(), "abc123"))
}

func TestService_List(t *testing.T) {
	mockClient := &mockIAMClient{
		GETFunc: func(ctx context.Context, url string, expectedResponseCode int, forceNewBearer bool) ([]byte, error) {
			assert.Equal(t, fmt.Sprintf("%s/env/v2/accounts/%s/environments", testEndpointURL, testAccountID), url)
			assert.Equal(t, 200, expectedResponseCode)
			return []byte(`{"data":[{"id":"abc123"},{"id":"def456"}]}`), nil
		},
	}

	service := createTestService(mockClient)
	stubs, err := service.List(t.Context())
	assert.NoError(t, err)
	assert.Len(t, stubs, 2)
	assert.Equal(t, "abc123", stubs[0].ID)
	assert.Equal(t, "def456", stubs[1].ID)
}

func TestService_ListPropagatesError(t *testing.T) {
	mockErr := errors.New("failed")
	mockClient := &mockIAMClient{
		GETFunc: func(ctx context.Context, url string, expectedResponseCode int, forceNewBearer bool) ([]byte, error) {
			return nil, mockErr
		},
	}

	service := createTestService(mockClient)
	_, err := service.List(t.Context())
	assert.EqualError(t, err, "failed")
}
