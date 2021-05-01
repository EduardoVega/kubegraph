package graph

import (
	"context"
	"fmt"
	"io"

	"strings"

	"k8s.io/client-go/dynamic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// Builder holds the information to build the graph
type Builder struct {
	Client    dynamic.Interface
	Namespace string
	Kind      string
	Name      string
	Out       io.Writer
	DotGraph  bool
	ObjData
}

// ObjData holds the object and related objects data
type ObjData struct {
	Obj             unstructured.Unstructured
	Hierarchy       string
	RelatedObjsData []ObjData
}

// NewBuilder returns a new builder struct
func NewBuilder(client dynamic.Interface, out io.Writer, dotGraph bool, namespace, kind, name string) *Builder {
	return &Builder{
		Client:    client,
		Out:       out,
		DotGraph:  dotGraph,
		Namespace: namespace,
		Kind:      kind,
		Name:      name,
		ObjData:   ObjData{},
	}
}

// Build gets all the information required to build the graph
func (b *Builder) Build() error {
	klog.V(1).Infoln("get objects to build the graph")

	o, err := getObject(b.Client, b.Namespace, b.Kind, b.Name)
	if err != nil {
		return err
	}
	b.ObjData.Obj = o
	b.ObjData.Hierarchy = ""

	r, err := getRelatedObjects(b.Client, []string{}, b.ObjData.Obj, b.Namespace)
	if err != nil {
		return err
	}
	b.ObjData.RelatedObjsData = r

	klog.V(4).Infof("object data JSON %s", ToJSON(b.ObjData))

	p := NewPrinter(b.ObjData, b.DotGraph, b.Out)
	p.Print()
	if err != nil {
		return err
	}

	return nil
}

// getObject returns the requested object
func getObject(client dynamic.Interface, namespace, kind, name string) (unstructured.Unstructured, error) {
	klog.V(1).Infof("get main object '%s'", kind)
	klog.V(2).Infof("get main object '%s' has finished", kind)

	gvr, err := getGroupVersionResource(kind)
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	ri := client.Resource(gvr).Namespace(namespace)

	obj, err := ri.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	return *obj, nil
}

// getRelatedObjects returns the list of upper and lower related objects
func getRelatedObjects(client dynamic.Interface, processedObjs []string, obj unstructured.Unstructured, namespace string) ([]ObjData, error) {
	klog.V(1).Infof("get related objects of kind '%s'", obj.GetKind())
	defer klog.V(2).Infof("get related objects of kind '%s' has finished", obj.GetKind())

	f := NewFilter()
	relatedObjs := []ObjData{}
	relatedKinds := getRelatedKinds(strings.ToLower(obj.GetKind()))
	processedObjs = append(processedObjs, strings.ToLower(obj.GetKind()))

	for hierarchy, kinds := range relatedKinds {
		klog.V(2).Infof("'%s' hierarchy '%s'", obj.GetKind(), hierarchy)

		for _, k := range kinds {
			klog.V(2).Infof("related object kind '%s'", k)
			// avoid calling the same obj multiple times
			if Contains(k, processedObjs) {
				klog.V(2).Infoln("skip")
				continue
			}

			gvr, err := getGroupVersionResource(k)
			if err != nil {
				return relatedObjs, err
			}

			ri := client.Resource(gvr).Namespace(namespace)

			objList, err := ri.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return relatedObjs, err
			}

			for _, o := range objList.Items {
				klog.V(2).Infof("filter related object '%s %s'", o.GetKind(), o.GetName())
				if f.FilterObj(obj, o) {
					klog.V(2).Infof("OK")
					r := ObjData{}
					r.Obj = o
					r.Hierarchy = hierarchy
					innerRelatedObjs, err := getRelatedObjects(client, processedObjs, o, namespace)
					if err != nil {
						return relatedObjs, err
					}
					r.RelatedObjsData = innerRelatedObjs
					relatedObjs = append(relatedObjs, r)
				}
			}
		}
	}

	return relatedObjs, nil
}

// getRelatedKinds returns a map of the related upper and lower kinds
func getRelatedKinds(kind string) map[string][]string {
	relatedkinds := map[string][]string{}

	switch kind {
	case "pod":
		relatedkinds = map[string][]string{
			"upper": {"service", "replicaset", "statefulset", "daemonset"},
			"lower": {},
		}
	case "service":
		relatedkinds = map[string][]string{
			"upper": {"ingress"},
			"lower": {"pod"},
		}
	case "ingress":
		relatedkinds = map[string][]string{
			"upper": {},
			"lower": {"service"},
		}
	case "replicaset":
		relatedkinds = map[string][]string{
			"upper": {"deployment"},
			"lower": {"pod"},
		}
	case "deployment":
		relatedkinds = map[string][]string{
			"upper": {},
			"lower": {"replicaset"},
		}
	case "statefulset", "daemonset":
		relatedkinds = map[string][]string{
			"upper": {},
			"lower": {"pod"},
		}
	}

	return relatedkinds
}

// getGroupVersionResource returns the correct group version resource struct
func getGroupVersionResource(kind string) (schema.GroupVersionResource, error) {
	switch kind {
	case "pod", "po":
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		}, nil
	case "service", "svc":
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "services",
		}, nil
	case "ingress", "ing":
		return schema.GroupVersionResource{
			Group:    "networking.k8s.io",
			Version:  "v1beta1",
			Resource: "ingresses",
		}, nil
	case "replicaset", "rs":
		return schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "replicasets",
		}, nil
	case "deployment", "deploy":
		return schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}, nil
	case "daemonset", "ds":
		return schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "daemonsets",
		}, nil
	case "statefulset", "sts":
		return schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "statefulsets",
		}, nil
	}

	return schema.GroupVersionResource{}, fmt.Errorf("kind '%s' not supported", kind)
}
