# readiness

Lightweight tool to wait for other Kubernetes resources to be ready.

## Usage

`Readiness` is provided as a docker container that can be used in a Kubernetes init container.
Example `Deployment`:

```yaml
spec:
  template:
    ...
    spec:
      ...
      initContainers:
        - name: my-app-readiness
          image: nexus3.onap.org:10001/onap/oom/readiness:7.0.0
          args:
          args:
            - '--namespace'
            - default
            - '--service-name'
            - my-app-postgres
```
