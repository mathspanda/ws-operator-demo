# ws-operator-demo
This is an operator demo for web server crd.

## How to build
``` shell
# build operator image
$ make operator-image

# build web server image
$ make ws-image
```

> Before building, please set environments: 
> * Registry 
> * RegistryAK
> * RegistrySK
> * RegistryPrefix

## How to use
### deploy operator
``` shell
$ helm install --name ws-demo-operator  --set resyncSeconds=150 ./helm/operator
```

### deploy WebServerCluster crd
``` shell
$ helm install --name ws-cluster-demo --set specData.replicas=4 --set specData.port=32241 ./helm/ws_cluster
```
Operator will listen crd ADD event, and create corresponding web server deployment and lb server:
``` shell
$ kubectl get po
NAME                               READY     STATUS    RESTARTS   AGE
ws-cluster-demo-2412484548-8rjjl   1/1       Running   0          2m
ws-cluster-demo-2412484548-nl9h0   1/1       Running   0          2m
ws-cluster-demo-2412484548-qg2mp   1/1       Running   0          2m
ws-cluster-demo-2412484548-s9wkk   1/1       Running   0          2m
ws-operator-demo-541433192-g2fs8   1/1       Running   0          5m
$ kubectl get svc
NAME              TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
ws-cluster-demo   LoadBalancer   10.96.175.38   <pending>     80:32241/TCP   3m
```

``` shell
$ for((i=1;i<=5;i++));do curl ${IP}:32241/ping;done
i am ws-cluster-demo-2412484548-s9wkk
i am ws-cluster-demo-2412484548-nl9h0
i am ws-cluster-demo-2412484548-s9wkk
i am ws-cluster-demo-2412484548-8rjjl
i am ws-cluster-demo-2412484548-qg2mp
```

### upgrade/delete WebServerCluster crd
```shell
$ helm upgrade --set XXX=XXX ws-cluster-demo ./helm/ws_cluster/
$ helm delete ws-cluster-demo --purge
```
