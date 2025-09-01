# provider-snowflake
A Crossplane provider built with Upjet for managing Snowflake resources.

---

## **Getting Started**

You can install this provider using the `crossplane` CLI. Just be sure to replace the image tag with the **latest release**:

```bash
crossplane xpkg install provider ghcr.io/allenkallz/provider-snowflake:v0.1.0
```

Alternatively, you can use declarative installation:
```
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-snowflake
spec:
  package: ghcr.io/allenkallz/provider-snowflake:v0.1.0
```

Notice that in this example Provider resource is referencing ControllerConfig with debug enabled.

You can see the API reference [here](https://doc.crds.dev/github.com/allenkallz/provider-snowflake).

## Developing

Run code-generation pipeline:
```console
go run cmd/generator/main.go "$PWD"
```

Run against a Kubernetes cluster:

```console
make run
```

Build, push, and install:

```console
make all
```

Build binary:

```console
make build
```

