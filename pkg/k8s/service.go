package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

type ServiceData struct {
	Name string

	Spec apiv1.ServiceSpec
}

type ServiceInterface interface {
	MakeConfig(*ServiceData) *apiv1.Service
	Create(*apiv1.Service) (*apiv1.Service, error)
	Delete(string, *metav1.DeleteOptions) error
	Update(*apiv1.Service) (*apiv1.Service, error)
	Get(string) (*apiv1.Service, error)
}

type services struct {
	client    v1.ServiceInterface
	namespace string
}

func NewService(kclient *kubernetes.Clientset, namespace string) ServiceInterface {
	return &services{
		client:    kclient.CoreV1().Services(namespace),
		namespace: namespace,
	}
}

func (s *services) MakeConfig(data *ServiceData) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.Name,
			Namespace: s.namespace,
		},
		Spec: data.Spec,
	}
}

func (s *services) Create(svcConfig *apiv1.Service) (*apiv1.Service, error) {
	return s.client.Create(svcConfig)
}

func (s *services) Delete(svcName string, options *metav1.DeleteOptions) error {
	return s.client.Delete(svcName, options)
}

func (s *services) Update(svcConfig *apiv1.Service) (*apiv1.Service, error) {
	return s.client.Update(svcConfig)
}

func (s *services) Get(svcName string) (*apiv1.Service, error) {
	return s.client.Get(svcName, metav1.GetOptions{})
}
