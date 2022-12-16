
```console
# 部署device model
kubectl apply -f crd/model.yaml
# 部署device
kubectl apply -f crd/instance.yaml

# 部署容器设备应用
kubectl apply -f deploy.yaml
```


```console
# 使用kubectl更新期望值
kubectl edit device light01

# 观察到light01的实际颜色也发生变化
kubectl get device light01 -ojson
```

```console
# 边缘节点本地文件中的数据也同样发生改变
cat /tmp/result.json
```