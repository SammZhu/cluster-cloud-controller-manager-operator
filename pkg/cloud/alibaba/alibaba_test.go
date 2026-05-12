package alibaba

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configv1 "github.com/openshift/api/config/v1"

	"github.com/openshift/cluster-cloud-controller-manager-operator/pkg/config"
)

func TestNewProviderAssets(t *testing.T) {
	tc := []struct {
		name       string
		config     config.OperatorConfig
		initErrMsg string
	}{
		{
			name:       "Empty config returns error",
			config:     config.OperatorConfig{},
			initErrMsg: "alibaba: missed images in config: CloudControllerManager: non zero value required",
		},
		{
			name: "Minimal allowed config",
			config: config.OperatorConfig{
				ImagesReference: config.ImagesReference{
					CloudControllerManagerAlibabaCloud: "registry.example.com/alibaba-cloud-controller-manager:latest",
				},
				PlatformStatus: &configv1.PlatformStatus{Type: configv1.AlibabaCloudPlatformType},
			},
		},
		{
			name: "Config with TLS settings",
			config: config.OperatorConfig{
				ImagesReference: config.ImagesReference{
					CloudControllerManagerAlibabaCloud: "registry.example.com/alibaba-cloud-controller-manager:latest",
				},
				PlatformStatus:  &configv1.PlatformStatus{Type: configv1.AlibabaCloudPlatformType},
				TLSCipherSuites: "TLS_AES_128_GCM_SHA256",
				TLSMinVersion:   "VersionTLS12",
			},
		},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			assets, err := NewProviderAssets(tc.config)
			if tc.initErrMsg != "" {
				assert.EqualError(t, err, tc.initErrMsg)
				return
			}
			require.NoError(t, err)

			resources := assets.GetRenderedResources()
			// deployment + serviceaccount + clusterrole + clusterrolebinding
			assert.Len(t, resources, 4)
		})
	}
}

func TestCloudConfigTransformer(t *testing.T) {
	tc := []struct {
		name        string
		source      string
		infra       *configv1.Infrastructure
		wantErr     bool
		wantContain []string
	}{
		{
			name:    "Nil infrastructure returns error",
			infra:   nil,
			wantErr: true,
		},
		{
			name: "Missing AlibabaCloud status returns error",
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						Type: configv1.AlibabaCloudPlatformType,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Valid config with region only",
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						Type: configv1.AlibabaCloudPlatformType,
						AlibabaCloud: &configv1.AlibabaCloudPlatformStatus{
							Region: "cn-hangzhou",
						},
					},
				},
			},
			wantContain: []string{"[Global]", "region=cn-hangzhou"},
		},
		{
			name: "Valid config with region and resourceGroupID",
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						Type: configv1.AlibabaCloudPlatformType,
						AlibabaCloud: &configv1.AlibabaCloudPlatformStatus{
							Region:          "cn-beijing",
							ResourceGroupID: "rg-abc123",
						},
					},
				},
			},
			wantContain: []string{"[Global]", "region=cn-beijing", "resourceGroupID=rg-abc123"},
		},
		{
			name:   "Source config is appended",
			source: "vpcid=vpc-12345\n",
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					PlatformStatus: &configv1.PlatformStatus{
						Type: configv1.AlibabaCloudPlatformType,
						AlibabaCloud: &configv1.AlibabaCloudPlatformStatus{
							Region: "cn-shanghai",
						},
					},
				},
			},
			wantContain: []string{"region=cn-shanghai", "vpcid=vpc-12345"},
		},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CloudConfigTransformer(tc.source, tc.infra, nil, nil)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			for _, want := range tc.wantContain {
				assert.Contains(t, result, want)
			}
		})
	}
}
