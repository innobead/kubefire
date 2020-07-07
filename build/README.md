# Build RootFS images

```
# make build-image-<distro> RELEASE=<release version>

make build-image-centos RELEASE=8
make build-image-ubuntu RELEASE=18.04
make build-image-ubuntu RELEASE=20.10
make build-image-opensuse-leap RELEASE=15.1
make build-image-sle15 RELEASE=15.1
```

# Build Kernel images

```
make build-kernel-4.19.125
make build-kernel-5.4.43
```
