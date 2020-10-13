# `kubegraph`: kubectl tool

[![Build Status](https://travis-ci.org/EduardoVega/kubegraph.svg?branch=master)](https://travis-ci.org/EduardoVega/kubegraph) [![Go Report Card](https://goreportcard.com/badge/github.com/EduardoVega/kubegraph)](https://goreportcard.com/report/github.com/EduardoVega/kubegraph)

kubegraph provides an easy way to visualize in your terminal, the relationship between k8s objects in a tree or dot graph.

## Using kubegraph

```
./kubegraph [OBJECT KIND] [OBJECT NAME]
```

Examples:
* Print a tree graph of the pod `my-pod` and its related Kubernetes objects.
    ```
    ./kubegraph pod my-pod
    ```
* Print a dot graph of the service `my-service` and its related Kubernetes objects. 
    ```
    ./kubegraph service my-service --dot
    ```
* Create an PNG Image using the output of a printed dot graph.
    ```
    ./kubegraph service my-service --dot | dot -Tpng > my-graph.png 
    ```

## Supported Kubernetes Object kinds

* pod
* service
* ingress
* replicaset
* deployment
* daemonset
* statefulset

## Installing kubegraph

### Pre-built Binaries

Download the latest binary from [releases](https://github.com/EduardoVega/kubegraph/releases)

Available OS/Architectures
* linux/amd64
* darwin/amd64
* windows/amd64

