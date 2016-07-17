# Letsencrypt Kubernetes

## WORK IN PROGRESS -- THIS TUTORIAL IS BROKEN!

The tutorial walks you through the process of obtaining TLS certificates from Letsencrypt and consuming them from an application.

## Generate the test certs

```
cd tls
```

```
cfssl gencert -initca ca-csr.json | cfssljson -bare ca
```

```
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -profile=kubernetes \
  server-csr.json | cfssljson -bare server
```

```
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -profile=kubernetes \
  server-updated-csr.json | cfssljson -bare server-updated
```

### Create the Kubernetes Secret

```
kubectl create secret tls tls-certs \
  --cert certs/server.pem \
  --key certs/server-key.pem
```

## Deploy the Application

```
kubectl create -f deployments/dynamic-certs.yaml
```

```
kubectl create -f services/dynamic-certs.yaml
```

## Test the Update process

```
kubectl delete secret tls-certs
```

```
kubectl create secret tls tls-certs \
  --cert certs/server-updated.pem \
  --key certs/server-updated-key.pem
```
