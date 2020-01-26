# go-kube-api (example service)

The service exposes an HTTP server and currently only works with an in-cluster config.

### Endpoints

#### POST /v1/rbac/enumerateBySubjectNames

Allows listing a namespace's RoleBindings based on their subject names either by exact value or a regular expression.

The endpoint requires a a `namespace` and one or more `subjectNames` either as alphanumeric strings (exact match)
or regular expressions which can be provided either as json or yaml depending on the `Content-Type` header.
The response will match the type of the request.

`Content-Type: application/json`

```json
{
  "namespace": "default",
  "subjectNames": [
    "subject1",
    "subject2",
    "subject[3,4]"
  ]
}
```

`Content-Type: application/x-yaml`

```yaml
namespace: default
subjectNames:
- subject1
- subject2
- subject[3,4]
```

## Building the binary

* Run `make build`. Service binary will be `./bin/go-kube-api`.

## Building the docker image

* Run `make docker`. Docker tag will be `go-kube-api:dev`.

## Running in-cluster (locally)

* Make sure you have a kubernetes with RBAC enabled.
* If you are running kubernetes from the docker mac app, you will have to
  first remove the global role that gives everyone admin access.

  ```sh
  kubectl delete ClusterRoleBinding docker-for-desktop-binding
  ```

* Build the docker image.
  
  ```sh
  make docker
  ```

* Apply the kubernetes manifests.
  
  ```sh
  kubectl apply -f deployment.yaml
  ```

* Check the service is running as expected.
  
  ```sh
  kubectl get deploy go-kube-api
  ```

* Port forward the service to your local machine.
  (or access the service any other way you can).

  ```sh
  kubectl port-forward service/go-kube-api 8080
  ```

* Check the health endpoint of the service.
  
  ```sh
  curl http://localhost:8080/healthz
  ```

* Add some sample roles and bindings.

  ```sh
  kubectl apply -f fixtures.yaml
  ```

* Make a request to retrieve role bindings by subject name.

  ```sh
  curl \
  -d '{"namespace":"default","subjectNames":["subject[3,4]"]}' \
  -H 'Content-Type: application/json' \
  http://localhost:8080/v1/rbac/enumerateBySubjectNames
  ```
