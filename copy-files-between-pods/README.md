

go build -o copyppp copy_pod_to_pod.go


# Apply ConfigMap and pods
kubectl apply -f configmap.yaml
kubectl apply -f pod-src.yaml
kubectl apply -f pod-dst.yaml

# Wait until pods are running
kubectl wait --for=condition=Ready pod/pod-src pod/pod-dst --timeout=60s

# Verify source file
kubectl exec pod-src -- cat /data/test.txt



./copyppp \
  --kubeconfig="$HOME/.kube/config" \
  --namespace=demo \
  --src-pod=ag1-0 \
  --src-container=mssql \
  --src-path=/var/opt/mssql/dbm_certificate.cer \
  --dst-pod=ag1-1 \
  --dst-container=mssql \
  --dst-dir=/var/opt/mssql








