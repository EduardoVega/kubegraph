package graph

import (
	"context"
	"fmt"
	"os"

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
	ObjData
}

// ObjData holds the object and related objects data
type ObjData struct {
	Obj             unstructured.Unstructured
	Hierarchy       string
	RelatedObjsData []ObjData
}

// NewBuilder returns a new builder struct
func NewBuilder(client dynamic.Interface, namespace, kind, name string) *Builder {
	return &Builder{
		Client:    client,
		Namespace: namespace,
		Kind:      kind,
		Name:      name,
		ObjData:   ObjData{},
	}
}

// Build gets all the information required to build the graph
func (b *Builder) Build() error {
	klog.V(1).Infoln("get objects to build the graph")

	o, err := b.GetObject(b.Namespace, b.Kind, b.Name)
	if err != nil {
		return err
	}
	b.ObjData.Obj = o
	b.ObjData.Hierarchy = ""

	r, err := b.GetRelatedObjects([]string{}, b.ObjData.Obj, b.Namespace)
	if err != nil {
		return err
	}
	b.ObjData.RelatedObjsData = r

	klog.V(4).Infof("object data JSON %s", ToJSON(b.ObjData))

	p := NewPrinter(b.ObjData, os.Stdout)
	p.Print("dot")
	if err != nil {
		return err
	}

	return nil
}

// GetObject returns the requested object
func (b *Builder) GetObject(namespace, kind, name string) (unstructured.Unstructured, error) {
	klog.V(1).Infof("get main object '%s'", kind)
	klog.V(2).Infof("get main object '%s' has finished", kind)

	var obj *unstructured.Unstructured

	gvr, err := b.GetGroupVersionResource(kind)
	if err != nil {
		return *obj, err
	}

	var ri dynamic.ResourceInterface
	ri = b.Client.Resource(gvr).Namespace(namespace)

	obj, err = ri.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return *obj, err
	}

	return *obj, nil
}

// GetRelatedObjects returns the list of upper and lower related objects
func (b *Builder) GetRelatedObjects(processedObjs []string, obj unstructured.Unstructured, namespace string) ([]ObjData, error) {
	klog.V(1).Infof("get related objects of kind '%s'", obj.GetKind())
	defer klog.V(2).Infof("get related objects of kind '%s' has finished", obj.GetKind())

	f := NewFilter()
	relatedObjs := []ObjData{}
	relatedKinds := b.GetRelatedKinds(strings.ToLower(obj.GetKind()))
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

			gvr, err := b.GetGroupVersionResource(k)
			if err != nil {
				return relatedObjs, err
			}

			var ri dynamic.ResourceInterface

			ri = b.Client.Resource(gvr).Namespace(namespace)

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
					innerRelatedObjs, err := b.GetRelatedObjects(processedObjs, o, namespace)
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

// GetRelatedKinds returns a map of the related upper and lower kinds
func (b *Builder) GetRelatedKinds(kind string) map[string][]string {
	relatedkinds := map[string][]string{}

	switch kind {
	case "pod":
		relatedkinds = map[string][]string{
			"upper": []string{"service", "replicaset", "statefulset"},
			"lower": []string{},
		}
	case "service":
		relatedkinds = map[string][]string{
			"upper": []string{"ingress"},
			"lower": []string{"pod"},
		}
	case "ingress":
		relatedkinds = map[string][]string{
			"upper": []string{},
			"lower": []string{"service"},
		}
	case "replicaset":
		relatedkinds = map[string][]string{
			"upper": []string{"deployment", "daemonset"},
			"lower": []string{"pod"},
		}
	case "deployment", "daemonset":
		relatedkinds = map[string][]string{
			"upper": []string{},
			"lower": []string{"replicaset"},
		}
	case "statefulset":
		relatedkinds = map[string][]string{
			"upper": []string{},
			"lower": []string{"pod"},
		}
	}

	return relatedkinds
}

// GetGroupVersionResource returns the correct group version resource struct
func (b *Builder) GetGroupVersionResource(kind string) (schema.GroupVersionResource, error) {
	var gvr schema.GroupVersionResource

	gvrMap := map[string]schema.GroupVersionResource{
		"pod": schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		},
		"service": schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "services",
		},
		"ingress": schema.GroupVersionResource{
			Group:    "networking.k8s.io",
			Version:  "v1beta1",
			Resource: "ingresses",
		},
		"replicaset": schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "replicasets",
		},
		"deployment": schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		},
		"daemonset": schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "daemonsets",
		},
		"statefulset": schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "statefulsets",
		},
	}

	if gvr, ok := gvrMap[kind]; ok {
		return gvr, nil
	}

	return gvr, fmt.Errorf("kind '%s' not found in supported GroupVersionResources", kind)
}
