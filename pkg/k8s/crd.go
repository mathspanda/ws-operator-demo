package k8s

import (
	"fmt"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

type CRD struct {
	Name          string
	Kind          string
	Plural        string
	Group         string
	Version       string
	Scope         apiextensionsv1beta1.ResourceScope
	Obj           runtime.Object
	ObjList       runtime.Object
	SchemeBuilder func(*runtime.Scheme) error
}

type CRDData struct {
	Name string
	Spec apiextensionsv1beta1.CustomResourceDefinitionSpec
}

func NewCRDData(config *CRD) *CRDData {
	return &CRDData{
		Name: config.Name,
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   config.Group,
			Version: config.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Kind:   config.Kind,
				Plural: config.Plural,
			},
		},
	}
}

type crds struct {
	client v1beta1.CustomResourceDefinitionInterface
}

type CRDRestClientConfig struct {
	KubeConfig *rest.Config
	CRD        *CRD
}

func NewCRD(clientset apiextensionsclient.Interface) CRDInterface {
	return &crds{
		client: clientset.ApiextensionsV1beta1().CustomResourceDefinitions(),
	}
}

type CRDInterface interface {
	MakeConfig(*CRDData) *apiextensionsv1beta1.CustomResourceDefinition
	Create(*apiextensionsv1beta1.CustomResourceDefinition) (*apiextensionsv1beta1.CustomResourceDefinition, error)
	Delete(string, *metav1.DeleteOptions) error

	NewRestClient(*CRDRestClientConfig) (*rest.RESTClient, *runtime.Scheme, error)
}

func (c *crds) MakeConfig(rawData *CRDData) *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: rawData.Name,
		},
		Spec: rawData.Spec,
	}
}

func (c *crds) Create(crdConfig *apiextensionsv1beta1.CustomResourceDefinition) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	crd, err := c.client.Create(crdConfig)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return c.client.Get(crdConfig.ObjectMeta.Name, metav1.GetOptions{})
		}
		return nil, err
	}

	err = c.waitCRDReady(crdConfig.ObjectMeta.Name)

	if err != nil {
		if deleteErr := c.client.Delete(crdConfig.ObjectMeta.Name, nil); deleteErr != nil {
			return nil, errors.NewAggregate([]error{err, deleteErr})
		}
		return nil, err
	}
	return crd, nil
}

func (c *crds) Delete(crdName string, options *metav1.DeleteOptions) error {
	return c.client.Delete(crdName, options)
}

func (c *crds) waitCRDReady(crdName string) error {
	err := wait.Poll(5*time.Second, 30*time.Second, func() (bool, error) {
		crd, err := c.client.Get(crdName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1beta1.Established:
				if cond.Status == apiextensionsv1beta1.ConditionTrue {
					return true, nil
				}
			case apiextensionsv1beta1.NamesAccepted:
				if cond.Status == apiextensionsv1beta1.ConditionFalse {
					return false, fmt.Errorf("Name conflict: %v", cond.Reason)
				}
			}
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("wait CRD created failed: %v", err)
	}
	return nil
}

func (c *crds) NewRestClient(config *CRDRestClientConfig) (*rest.RESTClient, *runtime.Scheme, error) {
	schemeBuilder := runtime.NewSchemeBuilder(config.CRD.SchemeBuilder)
	addToScheme := schemeBuilder.AddToScheme

	scheme := runtime.NewScheme()
	if err := addToScheme(scheme); err != nil {
		return nil, nil, err
	}

	cfg := *config.KubeConfig
	cfg.GroupVersion = &schema.GroupVersion{
		Group:   config.CRD.Group,
		Version: config.CRD.Version,
	}
	cfg.APIPath = "/apis"
	cfg.ContentType = runtime.ContentTypeJSON
	cfg.NegotiatedSerializer = serializer.DirectCodecFactory{
		CodecFactory: serializer.NewCodecFactory(scheme),
	}

	client, err := rest.RESTClientFor(&cfg)
	if err != nil {
		return nil, nil, err
	}

	return client, scheme, nil
}
