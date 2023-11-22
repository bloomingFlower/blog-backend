## 쿠버네티스 구성 초기화
1. kubeadm reset
2. sudo apt-get purge kubeadm kubectl kubelet
3. sudo rm -r /etc/kubernetes/
4. sudo rm -r /var/lib/etcd/
5. sudo rm -r /var/lib/kubelet
6. sudo rm -r ~/.kube
7. sudo systemctl daemon-reload
8. sudo rm -rf /etc/cni/net.d

## 쿠버네티스 설치
sudo apt-get update
apt-transport-https may be a dummy package; if so, you can skip that package
sudo apt-get install -y apt-transport-https ca-certificates curl gpg
```
1. ```shell
    curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
    ```

1. ```shell
    echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
    ```
1. ```shell
    sudo apt-get update
    sudo apt-get install -y kubelet kubeadm kubectl
    sudo apt-mark hold kubelet kubeadm kubectl
    ```

ystemctl daemon-reload
systemctl daemon-reload
systemctl restart kubelet
systemctl enable kubelet

kubeadm init --pod-network-cidr=192.168.0.0/16 --control-plane-endpoint=129.154.213.18
flannel --pod-network-cidr=10.244.0.0/16