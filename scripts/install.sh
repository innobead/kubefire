#!/usr/bin/env sh
# ref: https://ignite.readthedocs.io/en/stable/installation/

# 1. check virtualization capability

#$ lscpu | grep Virtualization
#Virtualization:      VT-x
#
#$ lsmod | grep kvm
#kvm_intel             200704  0
#kvm                   593920  1 kvm_intel

# 2. check and install docker or containerd

#yum install -y e2fsprogs openssh-clients
#which containerd || ( yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo && yum install -y containerd.io )
#    # Install containerd if it's not present

# 3. check and download CNI plugins

#export CNI_VERSION=v0.8.5
#export ARCH=$([ $(uname -m) = "x86_64" ] && echo amd64 || echo arm64)
#sudo mkdir -p /opt/cni/bin
#curl -sSL https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz | sudo tar -xz -C /opt/cni/bin

# 4. install ignite
#export VERSION=v0.7.0
#export GOARCH=$(go env GOARCH 2>/dev/null || echo "amd64")
#
#for binary in ignite ignited; do
#    echo "Installing ${binary}..."
#    curl -sfLo ${binary} https://github.com/weaveworks/ignite/releases/download/${VERSION}/${binary}-${GOARCH}
#    chmod +x ${binary}
#    sudo mv ${binary} /usr/local/bin
#done

# 5. verify
#$ ignite version
#Ignite version: version.Info{Major:"0", Minor:"6", GitVersion:"v0.7.0", GitCommit:"...", GitTreeState:"clean", BuildDate:"...", GoVersion:"...", Compiler:"gc", Platform:"linux/amd64"}
#Firecracker version: v0.18.1
#Runtime: containerd