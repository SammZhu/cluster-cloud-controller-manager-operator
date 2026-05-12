package alibaba

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
)

// CloudConfigTransformer transforms the cloud-config ConfigMap for AlibabaCloud.
// It injects region and resourceGroupID from the Infrastructure PlatformStatus into
// the [Global] section of the INI-format cloud config consumed by alibaba-cloud-controller-manager.
func CloudConfigTransformer(source string, infra *configv1.Infrastructure, _ *configv1.Network, _ featuregates.FeatureGate) (string, error) {
	if infra == nil || infra.Status.PlatformStatus == nil || infra.Status.PlatformStatus.AlibabaCloud == nil {
		return "", fmt.Errorf("alibabacloud platform status is not populated on infrastructure")
	}

	alibabaStatus := infra.Status.PlatformStatus.AlibabaCloud
	region := alibabaStatus.Region
	if region == "" {
		return "", fmt.Errorf("alibabacloud region is not set in platform status")
	}

	config := fmt.Sprintf("[Global]\nregion=%s\n", region)
	if alibabaStatus.ResourceGroupID != "" {
		config += fmt.Sprintf("resourceGroupID=%s\n", alibabaStatus.ResourceGroupID)
	}

	// Append caller-supplied source config (e.g. vpcId, vswitchId from install-config)
	if source != "" {
		config += "\n" + source
	}

	return config, nil
}
