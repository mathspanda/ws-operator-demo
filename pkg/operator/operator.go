package operator

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/mathspanda/ws-operator-demo/pkg/apis/demo.io/v1"
	"github.com/mathspanda/ws-operator-demo/pkg/controller"
	"github.com/mathspanda/ws-operator-demo/pkg/k8s"
)

type OperatorInterface interface {
	CreateCRD(*k8s.CRD) error
	DeleteCRD(string, *metav1.DeleteOptions) error
	// start controller to handle add/update/delete events
	WatchEvents(context.Context, *k8s.CRD) error
	Run(ctx context.Context, stopCh <-chan struct{}) error
}

type OperatorConfig struct {
	KubeConfigPath string
	WatchNamespace string
	ResyncPeriod   time.Duration
}

type WatchConfig struct {
	watchNamespace string
	resyncPeriod   time.Duration
	handlers       cache.ResourceEventHandler
}

type operator struct {
	kubeConfig *rest.Config
	// k8s clientset
	kubeClient *kubernetes.Clientset
	// apiextensions clientset
	aeClient *apiextensionsclient.Clientset

	crdI k8s.CRDInterface
	crd  *k8s.CRD

	watchConfig *WatchConfig

	logger *log.Entry
}

func NewOperator(config *OperatorConfig) (OperatorInterface, error) {
	kubeConfig, err := k8s.BuildKuberentesConfig(config.KubeConfigPath)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	aeClient, err := apiextensionsclient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	crd := &k8s.CRD{
		Name:          v1.CRDName,
		Kind:          v1.CRDKind,
		Plural:        v1.CRDPlural,
		Group:         v1.CRDGroup,
		Version:       v1.CRDVersion,
		Scope:         apiextensionsv1beta1.NamespaceScoped,
		Obj:           &v1.WebServerCluster{},
		ObjList:       &v1.WebServerClusterList{},
		SchemeBuilder: v1.AddKnownTypes,
	}

	watchConfig := &WatchConfig{
		resyncPeriod:   config.ResyncPeriod,
		watchNamespace: config.WatchNamespace,
		handlers: controller.NewWSController(&controller.WSControllerConfig{
			AEClient:   aeClient,
			KubeClient: kubeClient,
			Namespace:  config.WatchNamespace,
		}),
	}

	return &operator{
		kubeConfig:  kubeConfig,
		kubeClient:  kubeClient,
		aeClient:    aeClient,
		crdI:        k8s.NewCRD(aeClient),
		crd:         crd,
		watchConfig: watchConfig,
		logger:      log.WithField("app", "operator"),
	}, nil
}

func (o *operator) CreateCRD(crd *k8s.CRD) error {
	crdData := k8s.NewCRDData(crd)
	_, err := o.crdI.Create(o.crdI.MakeConfig(crdData))
	return err
}

func (o *operator) DeleteCRD(crdName string, options *metav1.DeleteOptions) error {
	return o.crdI.Delete(crdName, options)
}

func (o *operator) WatchEvents(ctx context.Context, crd *k8s.CRD) error {
	crdRestClientConfig := &k8s.CRDRestClientConfig{
		KubeConfig: o.kubeConfig,
		CRD:        o.crd,
	}
	crdRestClient, _, err := o.crdI.NewRestClient(crdRestClientConfig)
	if err != nil {
		return err
	}

	source := cache.NewListWatchFromClient(
		crdRestClient,
		o.crd.Plural,
		o.watchConfig.watchNamespace,
		fields.Everything(),
	)
	_, controller := cache.NewIndexerInformer(
		source,
		o.crd.Obj,
		o.watchConfig.resyncPeriod,
		o.watchConfig.handlers,
		cache.Indexers{},
	)

	go controller.Run(ctx.Done())

	return nil
}

func (o *operator) Run(ctx context.Context, stopCh <-chan struct{}) error {
	o.logger.Info("Begin to create crd.")
	if err := o.CreateCRD(o.crd); err != nil {
		return err
	}
	o.logger.Info("Successfully create crd.")

	o.logger.Info("Begin to watch events.")
	if err := o.WatchEvents(ctx, o.crd); err != nil {
		return err
	}

	<-stopCh
	return nil
}