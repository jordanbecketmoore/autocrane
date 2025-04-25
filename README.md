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
You can do the same for `destination` registries: 
```
apiVersion: image.autocrane.io/v1beta1
kind: CraneImage
spec: 
    source: 
        registry: docker.io
        credentialsSecret: dockerhub-creds
    destination: 
        registry: 123456789012.dkr.ecr.us-west-2.amazonaws.com
        credentialsSecret: ecr-creds
    image: 
        name: nginx
        tag: 1.21.6
```
#### Passthrough Cache
If you wish to route your image pull through a passthrough cache, you can specify the it under the `spec.passthroughCache`. For instance,
```
apiVersion: image.autocrane.io/v1beta1
kind: CraneImage
spec: 
    source: 
        registry: docker.io
    destination: 
        registry: 123456789012.dkr.ecr.us-west-2.amazonaws.com
        credentialsSecret: ecr-creds
    passthroughCache: 
        registry: my-passthrough-cache.com:5001
    image: 
        name: nginx
        tag: 1.21.6
```
will pull `my-passthrough-cache.com:5001/nginx:1.21.6`. 

You can specify a passthrough cache image prefix just as you can with source and destination. For instance,
```
apiVersion: image.autocrane.io/v1beta1
kind: CraneImage
spec: 
    source: 
        registry: docker.io
    destination: 
        registry: 123456789012.dkr.ecr.us-west-2.amazonaws.com
        credentialsSecret: ecr-creds
    passthroughCache: 
        registry: my-passthrough-cache.com:5001
        prefix: docker.io
    image: 
        name: nginx
        tag: 1.21.6
```
will pull `my-passthrough-cache.com:5001/docker.io/nginx:1.21.6`.

You can also specify the passthrough cache in a `CraneImagePolicy`, 
```
apiVersion: image.autocrane.io/v1beta1
kind: CraneImagePolicy
spec: 
    source: 
        registry: docker.io
    destination: 
        registry: 123456789012.dkr.ecr.us-west-2.amazonaws.com
        credentialsSecret: ecr-creds
    passthroughCache: 
        registry: my-passthrough-cache.com:5001
        prefix: docker.io
    imagePolicy: 
        name: 
            exact: nginx
        tag: 
            semver: ">=1.21.6"
```
*Note: The autocrane operator must have network access to the source registry in
order to get the list of image names and tags to apply an image policy.* 

### CraneImagePolicy
A CraneImagePolicy defines a set of CraneImage objects according to some matching
rule. The CraneImagePolicy resource controller checks source registry for images
that match the rule and deploys CraneImage objects for each matching image.
#### Tags
##### Semantic Version Constraints (semver)
A `CranImagePolicy` can match against image tags that conform to a semantic version constraint. Autocrane utilizes the 
[Masterminds/semver](https://github.com/Masterminds/semver) package to evaluate
image tags as semantic versions (where applicable) and check them against a
constraint provided in `spec.imagePolicy.tag.semver`. 

For example, the `CraneImagePolicy` defined here, 

```
apiVersion: image.autocrane.io/v1beta1
kind: CraneImagePolicy
metadata:
  name: ubuntu-semantic-version
spec:
  source:
    registry: docker.io
  destination:
    registry: my-private-registry.io
  imagePolicy:
    name: 
      exact: ubuntu
    tag: 
      semver: ">=20.0.0"
```

will match all ubuntu images that have semantic versions greater than 20.0.0. 

##### Regular Expressions (regex)
A CraneImagePolicy can match against image tags with a regex string. For example, if I wanted to generate CraneImages that matched DockerHub `ubuntu` images with tags that consisted of major and minor semantic versions, like `21.04`, `22.10`, `12.04`, I could create: 
```
apiVersion: image.autocrane.io/v1beta1
kind: CraneImagePolicy
metadata:
  name: ubuntu-minor-versions
spec:
  source:
    registry: docker.io
  destination:
    registry: my-private-registry.io
  imagePolicy:
    name: 
      exact: ubuntu
    tag: 
      regex: "^[0-9]+\\.[0-9]+$"
```
(Kubernetes requires that we double-escape the middle dot in the regex expression.)

This will generate CraneImage objects for all `ubuntu` images with tags that match the regex expression `^[0-9]+\.[0-9]+$`.



# TODOs 

## Features
- Loading registry credentials in Kubernetes on major cloud providers. 
- Public autocrane image metrics

## Proposed Sync Rules
- CalVer: Match an image based on its tag according to a calendar version constraint.

## QOL
- New documentation
- Cleaner statuses