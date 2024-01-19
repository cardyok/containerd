# Overlayfs snapshotter

aliyun internal overlayfs snapshotter, normal oci images will be converted on local machine to turboOCI image format. 

## Images Supported

Generally, we support three types of image, containerd will process these images in the following manners:

- **oci**: normal oci image, provided by all registries, will be converted to turboOCI image format
- **turboOCI**: turboOCI image, will be used directly
- **overlaybd format images**: dadi image, will be used directly and consumed by overlaybd

these images will put some requirements on the repository, to be more specific, we will describe steps to use for ACR: 

- **turboOCI**:

- *step 0*: you need an ACREE instance
- *step 1*: whitelist on ACR, contact ACR to whitelist your aliyun account.
- *step 2*: invoke [ACR openAPI](https://api.aliyun.com/api/cr/2018-12-01/CreateArtifactBuildRule?sdkStyle=old&params={%22Parameters%22:{%22ImageIndexOnly%22:%22true%22}}) to turn on turboOCI converter on your acr instance.
- *step 3*: select 仅索引模式 on your ACREE
- *step 4*: push an image
- *step 5*: use the image normally as you always do.
- *step 6*: start a container with your brand-new image.


- **overlaybd format images**:

- *step 0*: you need an ACREE instance
- *step 1*: select 完整模式 on your ACREE
- *step 2*: push an image
- *step 3*: add _containerd_accelerated as postfix to your original image tag.
- *step 4*: start a container with your brand-new image.

### Ref:
- [《TurboOCI本地镜像转换器》](https://aliyuque.antfin.com/storage/iqbu4a/ndffdla7egguhnir?singleDoc#)
- [《How Images Work on Aliyun》](https://aliyuque.antfin.com/pouchcontainer/wvz8ct/ip3nfc9inaiwi276?singleDoc#)
- [《How Overlaybd Integrates into AliyunNSM》](https://aliyuque.antfin.com/ansm/nsm/uqg2l0dlz0m0gwty#pZbqa)