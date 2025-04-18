# autocrane
A declarative implementation of the Crane image management tool for use in Kubernetes clusters

## kinds
### SyncImage
A SyncImage is a simple mapping of a container image in one registry to a copy
of that image in another registry. Source and destination image addresses are
provided, and the SyncImage controller ensures that the image at the source
address is mirrored to the destination address. 

```
kind: SyncImage
spec: 
    source: 
        registry: docker.io
    destination: 
        registry: 123456789012.dkr.ecr.us-west-2.amazonaws.com
    image: 
        name: nginx
        tag: 1.21.6
```


### SyncImageRule
A SyncImageRule defines a set of SyncImage objects according to some matching rule.