# 쿠버네티스 구성 초기화
1. kubeadm reset
2. sudo apt-get purge kubeadm kubectl kubelet
3. sudo rm -r /etc/kubernetes/
4. sudo rm -r /var/lib/etcd/
5. sudo rm -r /var/lib/kubelet
6. sudo rm -r ~/.kube
7. sudo systemctl daemon-reload
8. sudo rm -rf /etc/cni/net.d