# autocrane
A declarative implementation of the Crane image management tool for use in Kubernetes clusters

## Prerequisites
- Go version: 1.23.8
- Kubebuilder version: v4.5.2

## kinds
### CraneImage
A CraneImage is a simple mapping of a container image in one registry to a copy
of that image in another registry. Source and destination image addresses are
provided, and the CraneImage controller ensures that the image at the source
address is mirrored to the destination address. 

For example, the manifest below defines a CraneImage object mapping the DockerHub image `docker.io/nginx:1.21.6` to the same image in an ECR registry, `123456789012.dkr.ecr.us-west-2.amazonaws.com/nginx:1.21.6`. 
```
apiVersion: image.autocrane.io/v1beta1
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
`CraneImage` objects can take `docker-registry` type secrets as `source` and `destination` `credentialSecret` objects. 
For example, this Secret,
```
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: null
  name: dockerhub-creds
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: eyJhdXRocyI6eyJkb2NrZXIuaW8iOnsidXNlcm5hbWUiOiJ1c2VyIiwicGFzc3dvcmQiOiJwYXNzd29yZCIsImF1dGgiOiJkWE5sY2pwd1lYTnpkMjl5WkE9PSJ9fX0=
```
contains the docker config.json file: 
```
{
  "auths": {
    "docker.io": {
      "username": "user",
      "password": "password",
      "auth": "dXNlcjpwYXNzd29yZA=="
    }
  }
}
```
This contains dockerhub credentials, so I can use it to authenticate my source registry in the CraneImage below.
```
apiVersion: image.autocrane.io/v1beta1
kind: CraneImage
spec: 
    source: 
        registry: docker.io
        credentialsSecret: dockerhub-creds
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
- Prefix: Match an image by its address prefix.
- RegEx: Match an image using a regular expression.
- SemVer: Match an image based on its tag according to a semantic version constraint.
- CalVer: Match an image based on its tag according to a calendar version constraint.
