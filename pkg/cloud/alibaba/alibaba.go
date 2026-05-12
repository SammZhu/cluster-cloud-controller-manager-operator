package alibaba

import (
	"embed"
	"fmt"

	"github.com/asaskevich/govalidator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-cloud-controller-manager-operator/pkg/cloud/common"
	"github.com/openshift/cluster-cloud-controller-manager-operator/pkg/config"
)

const providerName = "alibaba"

var (
	//go:embed assets/*
	assetsFs embed.FS

	templates = []common.TemplateSource{
		{ReferenceObject: &appsv1.Deployment{}, EmbedFsPath: "assets/deployment.yaml"},
		{ReferenceObject: &corev1.ServiceAccount{}, EmbedFsPath: "assets/serviceaccount.yaml"},
		{ReferenceObject: &rbacv1.ClusterRole{}, EmbedFsPath: "assets/clusterrole.yaml"},
		{ReferenceObject: &rbacv1.ClusterRoleBinding{}, EmbedFsPath: "assets/clusterrolebinding.yaml"},
	}
)

type imagesReference struct {
	CloudControllerManager string `valid:"required"`
}

var templateValuesValidationMap = map[string]interface{}{
	"images":            "required",
	"cloudproviderName": "required,type(string)",
	"tlsCipherSuites":   "type(string)",
	"tlsMinVersion":     "type(string)",
}

type alibabaAssets struct {
	operatorConfig    config.OperatorConfig
	renderedResources []client.Object
}

func (a *alibabaAssets) GetRenderedResources() []client.Object {
	return a.renderedResources
}

func getTemplateValues(images *imagesReference, operatorConfig config.OperatorConfig) (common.TemplateValues, error) {
	values := common.TemplateValues{
		"images":            images,
		"cloudproviderName": operatorConfig.GetPlatformNameString(),
		"tlsCipherSuites":   operatorConfig.TLSCipherSuites,
		"tlsMinVersion":     operatorConfig.TLSMinVersion,
	}
	_, err := govalidator.ValidateMap(values, templateValuesValidationMap)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// NewProviderAssets returns CloudProviderAssets for AlibabaCloud platform.
func NewProviderAssets(config config.OperatorConfig) (common.CloudProviderAssets, error) {
	images := &imagesReference{
		CloudControllerManager: config.ImagesReference.CloudControllerManagerAlibabaCloud,
	}
	_, err := govalidator.ValidateStruct(images)
	if err != nil {
		return nil, fmt.Errorf("%s: missed images in config: %v", providerName, err)
	}
	assets := &alibabaAssets{
		operatorConfig: config,
	}
	objTemplates, err := common.ReadTemplates(assetsFs, templates)
	if err != nil {
		return nil, err
	}
	templateValues, err := getTemplateValues(images, config)
	if err != nil {
		return nil, err
	}
	assets.renderedResources, err = common.RenderTemplates(objTemplates, templateValues)
	if err != nil {
		return nil, err
	}
	return assets, nil
}
