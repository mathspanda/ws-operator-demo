# ws-operator-demo
It's an operator demo for web server crd. The operator will create CRD WebServerCluster first, and then listen
crd ADD/UPDATE/DELETE event.

ADD event: create web server deployment and lb service
UPDATE event: update web server deployment replica
DELETE event: delete corresponding deployment and service