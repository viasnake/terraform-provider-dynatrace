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
	"fmt"

	"github.com/dynatrace-oss/terraform-provider-dynatrace/dynatrace/api"
	ipallowlist "github.com/dynatrace-oss/terraform-provider-dynatrace/dynatrace/api/env/ipallowlist/settings"
	"github.com/dynatrace-oss/terraform-provider-dynatrace/dynatrace/api/iam"
	"github.com/dynatrace-oss/terraform-provider-dynatrace/dynatrace/rest"
	"github.com/dynatrace-oss/terraform-provider-dynatrace/dynatrace/settings"
)

type iamClientGetter interface {
	New(ctx context.Context) iam.IAMClient
}

type iamClientGetterImpl struct {
	clientID     string
	accountID    string
	clientSecret string
	tokenURL     string
	endpointURL  string
}

func (me *iamClientGetterImpl) ClientID() string {
	return me.clientID
}

func (me *iamClientGetterImpl) AccountID() string {
	return me.accountID
}

func (me *iamClientGetterImpl) ClientSecret() string {
	return me.clientSecret
}

func (me *iamClientGetterImpl) TokenURL() string {
	return me.tokenURL
}

func (me *iamClientGetterImpl) EndpointURL() string {
	return me.endpointURL
}

func (me *iamClientGetterImpl) New(ctx context.Context) iam.IAMClient {
	return iam.NewIAMClient(ctx, me)
}

type ServiceClient struct {
	iamClientGetter iamClientGetter
	accountID       string
	endpointURL     string
}

func NewService(credentials *rest.Credentials) settings.CRUDService[*ipallowlist.IPAllowlist] {
	return &ServiceClient{
		iamClientGetter: &iamClientGetterImpl{
			clientID:     credentials.IAM.ClientID,
			accountID:    credentials.IAM.AccountID,
			clientSecret: credentials.IAM.ClientSecret,
			tokenURL:     credentials.IAM.TokenURL,
			endpointURL:  credentials.IAM.EndpointURL,
		},
		accountID:   credentials.IAM.AccountID,
		endpointURL: credentials.IAM.EndpointURL,
	}
}

func Service(credentials *rest.Credentials) settings.CRUDService[*ipallowlist.IPAllowlist] {
	return NewService(credentials)
}

func (me *ServiceClient) SchemaID() string {
	return "accounts:environment:ip-allowlist"
}

func (me *ServiceClient) endpoint(environmentID string) string {
	return fmt.Sprintf("%s/env/v1/accounts/environments/%s/ip-allowlist", me.endpointURL, environmentID)
}

func (me *ServiceClient) List(ctx context.Context) (api.Stubs, error) {
	var environmentsResponse struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := iam.GET(me.iamClientGetter.New(ctx), ctx, fmt.Sprintf("%s/env/v2/accounts/%s/environments", me.endpointURL, me.accountID), 200, false, &environmentsResponse); err != nil {
		return nil, err
	}

	stubs := make(api.Stubs, 0, len(environmentsResponse.Data))
	for _, environment := range environmentsResponse.Data {
		if environment.ID == "" {
			continue
		}
		stubs = append(stubs, &api.Stub{ID: environment.ID, Name: environment.ID})
	}
	return stubs, nil
}

func (me *ServiceClient) Create(ctx context.Context, cfg *ipallowlist.IPAllowlist) (*api.Stub, error) {
	if _, err := me.iamClientGetter.New(ctx).PUT_MULTI_RESPONSE(ctx, me.endpoint(cfg.EnvironmentID), cfg, []int{200, 201, 204}, false); err != nil {
		return nil, err
	}
	return &api.Stub{ID: cfg.EnvironmentID, Name: cfg.EnvironmentID}, nil
}

func (me *ServiceClient) Get(ctx context.Context, id string, cfg *ipallowlist.IPAllowlist) error {
	if err := iam.GET(me.iamClientGetter.New(ctx), ctx, me.endpoint(id), 200, false, cfg); err != nil {
		return err
	}
	cfg.EnvironmentID = id
	if cfg.Allowlist == nil {
		cfg.Allowlist = []ipallowlist.AllowlistEntry{}
	}
	return nil
}

func (me *ServiceClient) Update(ctx context.Context, id string, cfg *ipallowlist.IPAllowlist) error {
	environmentID := id
	if cfg.EnvironmentID != "" {
		environmentID = cfg.EnvironmentID
	}
	_, err := me.iamClientGetter.New(ctx).PUT_MULTI_RESPONSE(ctx, me.endpoint(environmentID), cfg, []int{200, 201, 204}, false)
	return err
}

func (me *ServiceClient) Delete(ctx context.Context, id string) error {
	disabled := &ipallowlist.IPAllowlist{
		Enabled:              false,
		AllowWebhookOverride: false,
		Allowlist:            []ipallowlist.AllowlistEntry{},
	}
	_, err := me.iamClientGetter.New(ctx).PUT_MULTI_RESPONSE(ctx, me.endpoint(id), disabled, []int{200, 201, 204}, false)
	return err
}
