Besides, the prebuilt `weaveworks/ignite-kernel` kernel images, you can build customized ones as below steps.

# Build customized kernel image

1. Clone https://github.com/weaveworks/ignite
2. Go to images/kernel
3. Update the kernel config of specified version if need (like enable apparmor security module)
4. Build a kernel image or all 

```
# build all kernel images
make

# build one kernel image
make build-4.19.125
make build-5.4.43
```