package controller

import (
	"reflect"

	log "github.com/sirupsen/logrus"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/util/workqueue"

	"github.com/mathspanda/ws-operator-demo/pkg/apis/demo.io/v1"
	"github.com/mathspanda/ws-operator-demo/pkg/k8s"
)

const (
	maxRetries = 15
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

	queue workqueue.RateLimitingInterface

	logger *log.Entry
}

func NewWSController(config *WSControllerConfig) *WSController {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
		"ws-cluster-queue")

	return &WSController{
		aeClient:   config.AEClient,
		kubeClient: config.KubeClient,
		deployI:    k8s.NewDeployment(config.KubeClient, config.Namespace),
		svcI:       k8s.NewService(config.KubeClient, config.Namespace),
		queue:      queue,
		logger:     log.WithField("service", "controller"),
	}
}

func (w *WSController) Worker() {
	for w.processNextWorkItem() {
	}
}

func (w *WSController) processNextWorkItem() bool {
	task, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(task)

	err := w.sync(task)
	w.handleErr(err, task)

	return true
}

func (w *WSController) handleErr(err error, task interface{}) {
	if err == nil {
		w.queue.Forget(task)
		return
	}

	if w.queue.NumRequeues(task) < maxRetries {
		w.logger.Infof("Error syncing CRD: %v, %v", task, err)
		w.queue.AddRateLimited(task)
		return
	}

	utilruntime.HandleError(err)
	w.logger.Errorf("Dropping CRD %q out of the queue: %v", task, err)
	w.queue.Forget(task)
}

func (w *WSController) enqueueTask(taskType TaskType, ws *v1.WebServerCluster) {
	w.queue.Add(&CRDTask{
		CRDTaskType: taskType,
		CRDObj:      ws,
	})
}

func (w *WSController) sync(task interface{}) error {
	crdTask := task.(*CRDTask)
	wsCluster := crdTask.CRDObj.(*v1.WebServerCluster)

	var err error
	switch crdTask.CRDTaskType {
	case TaskTypeAdd:
		err = w.createWebServerCluster(wsCluster)
	case TaskTypeUpdate:
		err = w.updateWebServerCluster(wsCluster)
	case TaskTypeDelete:
		err = w.deleteWebServerCluster(wsCluster)
	}

	return err
}

func (w *WSController) OnAdd(obj interface{}) {
	wsCluster := obj.(*v1.WebServerCluster)
	w.enqueueTask(TaskTypeAdd, wsCluster)
}

func (w *WSController) OnUpdate(oldObj, newObj interface{}) {
	oldWSCluster := oldObj.(*v1.WebServerCluster)
	newWSCluster := newObj.(*v1.WebServerCluster)

	if !reflect.DeepEqual(oldWSCluster.Spec, newWSCluster.Spec) {
		w.enqueueTask(TaskTypeUpdate, newWSCluster)
	}
}

func (w *WSController) OnDelete(obj interface{}) {
	wsCluster := obj.(*v1.WebServerCluster)
	w.enqueueTask(TaskTypeDelete, wsCluster)
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
	w.logger.Infof("Successfully delete web server cluster %s", ws.ObjectMeta.Name)
	return nil
}

func (w *WSController) updateWebServerCluster(ws *v1.WebServerCluster) error {
	wsDeployData := w.newWebServerClusterDeploymentData(ws)
	if _, err := w.deployI.Update(w.deployI.MakeConfig(wsDeployData)); err != nil {
		return err
	}
	w.logger.Infof("Successfully update web server cluster %s", ws.ObjectMeta.Name)
	return nil
}

func (w *WSController) createWebServerCluster(ws *v1.WebServerCluster) error {
	wsDeployData := w.newWebServerClusterDeploymentData(ws)
	_, err := w.deployI.Create(w.deployI.MakeConfig(wsDeployData))
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	wsServiceData := w.newWebServerClusterServiceData(ws)
	_, err = w.svcI.Create(w.svcI.MakeConfig(wsServiceData))
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	w.logger.Infof("Successfully create web server cluster %s", ws.ObjectMeta.Name)
	return nil
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
