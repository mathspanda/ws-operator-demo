package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type DeploymentData struct {
	Name string

	Spec extensionsv1beta1.DeploymentSpec
}

type DeploymentInterface interface {
	MakeConfig(*DeploymentData) *extensionsv1beta1.Deployment
	Create(*extensionsv1beta1.Deployment) (*extensionsv1beta1.Deployment, error)
	Delete(string, *metav1.DeleteOptions) error
	Update(*extensionsv1beta1.Deployment) (*extensionsv1beta1.Deployment, error)
}

type deployments struct {
	client    v1beta1.DeploymentInterface
	namespace string
}

func NewDeployment(kclient *kubernetes.Clientset, namespace string) DeploymentInterface {
	return &deployments{
		client:    kclient.ExtensionsV1beta1().Deployments(namespace),
		namespace: namespace,
	}
}

func (d *deployments) MakeConfig(data *DeploymentData) *extensionsv1beta1.Deployment {
	return &extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Name,
			Namespace: d.namespace,
		},
		Spec: data.Spec,
	}
}

func (d *deployments) Create(deploy *extensionsv1beta1.Deployment) (*extensionsv1beta1.Deployment, error) {
	return d.client.Create(deploy)
}

func (d *deployments) Delete(deployName string, options *metav1.DeleteOptions) error {
	return d.client.Delete(deployName, options)
}

func (d *deployments) Update(deploy *extensionsv1beta1.Deployment) (*extensionsv1beta1.Deployment, error) {
	return d.client.Update(deploy)
}