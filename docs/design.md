# Kong Operator design doc

Kong operator will act on two CRDs:
* `KongService`
* `KongRoute`

Kong operator will perform theses actions on Kubernetes resources:
* `KongService` *Update - Get - Watch*
* `KongRoute` *Update - Get - Watch*
* `Service` *Get - Watch*

## KongService

`KongService` is the kubernetes abstraction of kong service introduced in kong `0.13`. It's use to abstract kubernetes service upstream

`KongService` custom ressource definition:
```yaml
apiVersion: k8s.cdiscount.com/v1alpha1
kind: KongService
metadata:
  name: app-service
spec:
  serviceName: myservice
  path: /api
  retries: 10
  connectTimeout: 60000
  writeTimeout: 60000
  readTimeout: 60000
status:
  [...]
```

### KongService Spec

`KongService` spec will be base on the following parameters

 Attributes        | Description           |Default| Optional  |
|-------------|-------------|-----|-|
| serviceName | Reference existing kubernetes service by name || No |
| path      | The path to be used in requests to the upstream server.|/| Yes |
|retries|The number of retries to execute upon failure to proxy.|5|Yes|
|connectTimeout|The timeout in milliseconds for establishing a connection to the upstream server.|60000|Yes|
|writeTimeout|The timeout in milliseconds between two successive write operations for transmitting a request to the upstream server.|60000|Yes|
|readTimeout|The timeout in milliseconds between two successive read operations for transmitting a request to the upstream server.|60000|Yes|

### KongService Status

`KongService` status will register these informations

Attributes        | Description|
|-------------|-------------|
|kongStatus| Kong status, possible value ["Registered", "Unregistered"]|
|kongId| Reference kong id|
|url|composition of host + port + path|
|createdAt|Creation timestamp|
|updatedAt|Update timestamp|

```yaml
apiVersion: k8s.cdiscount.com/v1alpha1
kind: KongService
metadata:
  name: app-service
spec:
 [...]
status:
  kongStatus: "Registered"
  kongId: "4e13f54a-bbf1-47a8-8777-255fed7116f2"
  url: api.com:80/api
  createdAt: 1488869076800
  updatedAt: 1488869076800
```

### KongService Lifecycle

#### Create
When new `KongService` is created, it will register a new Kong service based on `Kubernetes service` information and `status.kongStatus` will be changed from `Unregistered` to `Registered`. If `Kubernetes service` doesn't exist, No error will be raised but an event will be register.

#### Update
when `KongService` is updated, it will update Kong service. If exposed `kubernetes service` change, Kong Operator will update Kong with new reference automatically.

#### Delete
Deleting a `KongService` will delete kong service reference.

## KongRoute
`KongRoute` is the kubernetes abstraction to represent `Kong` route. Each `KongRoute` is associated with a `KongService` and a `KongService` may have multiple `KongRoute` associated.

`KongRoute` custom resource definition:
```yaml
apiVersion: k8s.cdiscount.com/v1alpha1
kind: KongRoute
metadata:
  name: app-route
spec:
  kongServiceName: app-service
  protocols:
  - http
  - https
  methods:
  - GET
  - POST
  hosts:
  - example.com
  paths:
  - /api
  stripPath: true
  preserveHost: false
status:
  [...]
```

### KongRoute Spec

`KongRoute` spec will be base on the following parameters

 Attributes        | Description           |Default| Optional  |
|-------------|-------------|-----|-|
| KongServiceName | Reference existing kubernetes kong service by name || No |
| protocols      |  List of the protocols this Route should allow.|["http", "https"]| Yes |
|methods|A list of HTTP methods that match this Route.| ["GET"]| Yes |
|hosts|A list of domain names that match this Route||No|
|paths|A list of paths that match this Route.||No|
|stripPath|When matching a Route via one of the paths, strip the matching prefix from the upstream request URL|true|Yes|
|preserveHost|When matching a Route via one of the hosts domain names, use the request Host header in the upstream request headers|false|Yes|

### KongRoute Status

`KongRoute` status will register these informations:

Attributes        | Description|
|-------------|-------------|
|kongStatus| Kong status, possible value ["Registered", "Unregistered"]|
|kongId| Reference kong id|
|serviceIdRef| Reference kong service id|
|url|composition of host + port + path|
|createdAt|Creation timestamp|
|updatedAt|Update timestamp|

```yaml
apiVersion: k8s.cdiscount.com/v1alpha1
kind: KongRoute
metadata:
  name: app-route
spec:
 [...]
status:
  kongStatus: "Register"
  kongId: "4e13f54a-bbf1-47a8-8777-255fed7116f2"
  serviceIdRef: "22108377-8f26-4c0e-bd9e-2962c1d6b0e6"
  createdAt: 1488869076800
  updatedAt: 1488869076800
```

### KongRoute Lifecycle

#### Create
When new `Kongroute` is created, it will register a new Kong route based on `KongService` informations and `status.kongStatus` will be changed from `Unregistered` to `Registered`. If `KongService` doesn't exist, No error will be raised but an event will be register.

#### Update
when `KongRoute` is updated, it will update Kong Route. If exposed `KongService` change, Kong Operator will update Kong with new reference automatically.

#### Delete
Deleting a `KongRoute` will delete kong route reference.

## Futur improvement

Instead of acting on kubernetes `service` and reference them in Kong, Kong Operator could reference directly `endpoint` to increase performance.
