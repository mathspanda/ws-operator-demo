package controller

import (
	log "github.com/sirupsen/logrus"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/mathspanda/ws-operator-demo/pkg/apis/demo.io/v1"
	"github.com/mathspanda/ws-operator-demo/pkg/k8s"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
)

type WSControllerConfig struct {
	AEClient   *apiextensionsclient.Clientset
	KubeClient *kubernetes.Clientset
	Namespace  string
}

type WSController struct {
	aeClient   *apiextensionsclient.Clientset
	kubeClient *kubernetes.Clientset

	deployI k8s.DeploymentInterface
	svcI    k8s.ServiceInterface

	logger *log.Entry
}

func NewWSController(config *WSControllerConfig) *WSController {
	return &WSController{
		aeClient:   config.AEClient,
		kubeClient: config.KubeClient,
		deployI:    k8s.NewDeployment(config.KubeClient, config.Namespace),
		svcI:       k8s.NewService(config.KubeClient, config.Namespace),

		logger: log.WithField("service", "controller"),
	}
}

func (w *WSController) OnAdd(obj interface{}) {
	wsCluster := obj.(*v1.WebServerCluster)
	err := w.createWebServerCluster(wsCluster)
	if err != nil {
		w.logger.Errorf("Failed to create web server cluster %s: %v", wsCluster.ObjectMeta.Name, err)
		return
	}
	w.logger.Infof("Successfully create web server cluster %s", wsCluster.ObjectMeta.Name)
}

func (w *WSController) OnUpdate(oldObj, newObj interface{}) {
	oldWSCluster := oldObj.(*v1.WebServerCluster)
	newWSCluster := newObj.(*v1.WebServerCluster)

	if !reflect.DeepEqual(oldWSCluster.Spec, newWSCluster.Spec) {
		if err := w.updateWebServerCluster(newWSCluster); err != nil {
			w.logger.Errorf("Failed to update web server cluster %s: %v", oldWSCluster.ObjectMeta.Name, err)
			return
		}
		w.logger.Infof("Successfully update web server cluster %s", oldWSCluster.ObjectMeta.Name)
	}
}

func (w *WSController) OnDelete(obj interface{}) {
	wsCluster := obj.(*v1.WebServerCluster)
	err := w.deleteWebServerCluster(wsCluster)
	if err != nil {
		w.logger.Errorf("Failed to delete web server cluster %s: %v", wsCluster.ObjectMeta.Name, err)
		return
	}
	w.logger.Infof("Successfully delete web server cluster %s", wsCluster.ObjectMeta.Name)
}

func (w *WSController) deleteWebServerCluster(ws *v1.WebServerCluster) error {
	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := w.deployI.Delete(ws.ObjectMeta.Name, deleteOptions); err != nil {
		return err
	}
	if err := w.svcI.Delete(ws.ObjectMeta.Name, nil); err != nil {
		return err
	}
	return nil
}

func (w *WSController) updateWebServerCluster(ws *v1.WebServerCluster) error {
	wsDeployData := w.newWebServerClusterDeploymentData(ws)
	_, err := w.deployI.Update(w.deployI.MakeConfig(wsDeployData))
	return err
}

func (w *WSController) createWebServerCluster(ws *v1.WebServerCluster) error {
	wsDeployData := w.newWebServerClusterDeploymentData(ws)
	_, err := w.deployI.Create(w.deployI.MakeConfig(wsDeployData))
	if err != nil {
		return err
	}

	wsServiceData := w.newWebServerClusterServiceData(ws)
	_, err = w.svcI.Create(w.svcI.MakeConfig(wsServiceData))
	return err
}

func (w *WSController) newWebServerClusterDeploymentData(ws *v1.WebServerCluster) *k8s.DeploymentData {
	return &k8s.DeploymentData{
		Name: ws.ObjectMeta.Name,
		Spec: extensionsv1beta1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "ws-cluster-" + ws.ObjectMeta.Name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "ws-cluster-" + ws.ObjectMeta.Name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "ws-" + ws.ObjectMeta.Name,
							Image: ws.Spec.Image,
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
			Replicas: ws.Spec.Replicas,
		},
	}
}

func (w *WSController) newWebServerClusterServiceData(ws *v1.WebServerCluster) *k8s.ServiceData {
	return &k8s.ServiceData{
		Name: ws.ObjectMeta.Name,
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"app": "ws-cluster-" + ws.ObjectMeta.Name,
			},
			Ports: []apiv1.ServicePort{
				{
					TargetPort: intstr.FromInt(80),
					NodePort:   ws.Spec.ServicePort,
					Port:       80,
				},
			},
			Type: "LoadBalancer",
		},
	}
}
