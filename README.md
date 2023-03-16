
# Kuberenetes Volume Usage Prometheus Metric (kubelet way)

## Deploy

```bash
kubectl apply -f sample-deploy.yaml
# give access to get node state
kubectl create clusterrolebinding exporter-cluster-role --clusterrole=cluster-admin --serviceaccount=kube-system:default
# test
curl volume-usage-exporter.kube-system.svc.cluster.local:9001/metrics
```

## Metrics

| Metric name | Metric type | Labels |
|-------------|-------------|-------------|
|kubelet_volume_stats_capacity_bytes|Gauge|namespace=\<persistentvolumeclaim-namespace\> <br/> persistentvolumeclaim=\<persistentvolumeclaim-name\>| 
|kubelet_volume_stats_available_bytes|Gauge|namespace=\<persistentvolumeclaim-namespace\> <br/> persistentvolumeclaim=\<persistentvolumeclaim-name\>| 
|kubelet_volume_stats_used_bytes|Gauge|namespace=\<persistentvolumeclaim-namespace\> <br/> persistentvolumeclaim=\<persistentvolumeclaim-name\>| 
|kubelet_volume_stats_inodes|Gauge|namespace=\<persistentvolumeclaim-namespace\> <br/> persistentvolumeclaim=\<persistentvolumeclaim-name\>| 
|kubelet_volume_stats_inodes_free|Gauge|namespace=\<persistentvolumeclaim-namespace\> <br/> persistentvolumeclaim=\<persistentvolumeclaim-name\>| 
|kubelet_volume_stats_inodes_used|Gauge|namespace=\<persistentvolumeclaim-namespace\> <br/> persistentvolumeclaim=\<persistentvolumeclaim-name\>| 

## Build
```bash
docker build --tag kubernetes-volume-usage-prometheus-metric:local . -f Dockerfile
# For multi plateform 
# docker buildx build --push --platform linux/arm/v7,linux/arm64,linux/amd64 --tag medinvention/kubernetes-volume-usage-prometheus-metric:0.0.1 . -f Dockerfile
```


### References

- https://github.com/kubernetes/kubernetes/pull/51553
- https://github.com/kubernetes/community/pull/855