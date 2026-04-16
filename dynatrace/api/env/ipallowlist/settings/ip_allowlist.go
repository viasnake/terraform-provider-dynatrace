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
	"github.com/dynatrace-oss/terraform-provider-dynatrace/terraform/hcl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type IPAllowlist struct {
	EnvironmentID        string           `json:"-"`
	Enabled              bool             `json:"enabled"`
	AllowWebhookOverride bool             `json:"allowWebhookOverride"`
	Allowlist            []AllowlistEntry `json:"allowlist"`
}

type AllowlistEntry struct {
	Name    string `json:"name"`
	IPRange string `json:"ipRange"`
}

func (me *IPAllowlist) Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"environment_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The environment ID (`https://<environment-id>.live.dynatrace.com`) this allowlist applies to.",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Whether IP allowlisting is enabled.",
		},
		"allow_webhook_override": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether webhook calls are allowed to bypass IP allowlist restrictions.",
		},
		"allowlist": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "Allowed IP ranges.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A display name for this allowlist entry.",
				},
				"ip_range": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "IP range in CIDR notation.",
				},
			}},
		},
	}
}

func (me *IPAllowlist) MarshalHCL(properties hcl.Properties) error {
	return properties.EncodeAll(map[string]any{
		"environment_id":         me.EnvironmentID,
		"enabled":                me.Enabled,
		"allow_webhook_override": me.AllowWebhookOverride,
		"allowlist":              me.Allowlist,
	})
}

func (me *IPAllowlist) UnmarshalHCL(decoder hcl.Decoder) error {
	if err := decoder.DecodeAll(map[string]any{
		"environment_id":         &me.EnvironmentID,
		"enabled":                &me.Enabled,
		"allow_webhook_override": &me.AllowWebhookOverride,
		"allowlist":              &me.Allowlist,
	}); err != nil {
		return err
	}
	if me.Allowlist == nil {
		me.Allowlist = []AllowlistEntry{}
	}
	return nil
}

func (me AllowlistEntry) MarshalHCL(properties hcl.Properties) error {
	return properties.EncodeAll(map[string]any{
		"name":     me.Name,
		"ip_range": me.IPRange,
	})
}

func (me *AllowlistEntry) UnmarshalHCL(decoder hcl.Decoder) error {
	return decoder.DecodeAll(map[string]any{
		"name":     &me.Name,
		"ip_range": &me.IPRange,
	})
}
