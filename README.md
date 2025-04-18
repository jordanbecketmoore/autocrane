# autocrane
A declarative implementation of the Crane image management tool for use in Kubernetes clusters

## Prerequisites
- Go version: 
- Kubebuilder version: 

## kinds
### CraneImage
A CraneImage is a simple mapping of a container image in one registry to a copy
of that image in another registry. Source and destination image addresses are
provided, and the CraneImage controller ensures that the image at the source
address is mirrored to the destination address. 

For example, the manifest below defines a CraneImage object mapping the DockerHub image `docker.io/nginx:1.21.6` to the same image in an ECR registry, `123456789012.dkr.ecr.us-west-2.amazonaws.com/nginx:1.21.6`. 
```
apiVersion: autocrane.io/v1beta1
kind: CraneImage
spec: 
    source: 
        registry: docker.io
    destination: 
        registry: 123456789012.dkr.ecr.us-west-2.amazonaws.com
    image: 
        name: nginx
        tag: 1.21.6
```


### CraneImageRule
A CraneImageRule defines a set of CraneImage objects according to some matching
rule. The CraneImageRule resource controller checks source registry for images
that match the rule and deploys CraneImage objects for each matching image.

#### Proposed Sync Rules
- Prefix: Match an image its address prefix
- RegEx: Match an image from a regular expression. 
- SemVer: Match an image based on its tag according to a semantic version constraint. 

### CraneSecret
A CraneSecret contains credentials for accessing a remote image registry.