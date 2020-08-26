package graph

import (
	"context"
	"fmt"

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

// Graph holds the objects of the graph
// type Graph struct {
// 	Obj         unstructured.Unstructured
// 	RelatedObjs []RelatedObj
// }

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
	// ownerObjs, err := GetOwnerObjects(b.Client, b.Graph.Obj, b.Namespace)
	// if err != nil {
	// 	return err
	// }
	// klog.V(3).Infof("ownerObjects %s", GetJSON(ownerObjs))
	// b.Graph.OwnerObjs = ownerObjs

	// err = GetRelatedObjects(b.Client, b.Graph, b.Namespace, b.Kind, b.Name)
	// if err != nil {
	// 	return err
	// }

	// dotg, err := BuildDOTGraph(b.Graph)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(dotg)

	return nil
}

// GetObject returns the requested object
func (b *Builder) GetObject(namespace, kind, name string) (unstructured.Unstructured, error) {
	klog.V(1).Infof("get main object '%s'", kind)
	defer klog.V(2).Infof("get main object '%s' has finished", kind)

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

// OwnerObj holds the object obtained from ownerRefences
// type OwnerObj struct {
// 	Obj       unstructured.Unstructured
// 	OwnerObjs []OwnerObj
// }

// GetOwnerObjects returns the owner objects
// func GetOwnerObjects(client dynamic.Interface, obj unstructured.Unstructured, namespace string) ([]OwnerObj, error) {
// 	klog.V(1).Infoln("get owner objects")

// 	ownerObjs := []OwnerObj{}

// 	for _, ownerReference := range obj.GetOwnerReferences() {
// 		gvr, err := GetGroupVersionResource(strings.ToLower(ownerReference.Kind))
// 		if err != nil {
// 			return ownerObjs, nil // Not supported GVR
// 		}

// 		var ri dynamic.ResourceInterface
// 		ri = client.Resource(gvr).Namespace(namespace)

// 		ownerObj, err := ri.Get(context.TODO(), ownerReference.Name, metav1.GetOptions{})
// 		if err != nil {
// 			return ownerObjs, err
// 		}

// 		ob := OwnerObj{}
// 		ob.Obj = *ownerObj

// 		innerOwnerObjs, err := GetOwnerObjects(client, *ownerObj, namespace)
// 		if err != nil {
// 			return ownerObjs, err
// 		}

// 		ob.OwnerObjs = innerOwnerObjs
// 		ownerObjs = append(ownerObjs, ob)
// 	}

// 	return ownerObjs, nil
// }

// BuildDOTGraph returns a DOT graph populated with obtained k8s objects
// func BuildDOTGraph(graph *Graph) (string, error) {
// 	klog.V(1).Infoln("build the dot graph")

// 	dotGraph := gographviz.NewGraph()

// 	err := dotGraph.SetName("G")
// 	if err != nil {
// 		return "", err
// 	}

// 	err = dotGraph.SetDir(true)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Add pod nodes
// 	for _, pod := range graph.Pods {

// 		err = dotGraph.AddNode("G", GetPrettyString(pod.Name), map[string]string{"label": "\"pod: " + pod.Name + "\""})
// 		if err != nil {
// 			return "", err
// 		}

// 		for _, container := range pod.Containers {
// 			err = dotGraph.AddNode("G", GetPrettyString(pod.Name+container), map[string]string{"label": "\"container: " + container + "\""})
// 			if err != nil {
// 				return "", err
// 			}

// 			dotGraph.AddEdge(GetPrettyString(pod.Name), GetPrettyString(pod.Name+container), true, nil)
// 		}

// 		for _, initContainer := range pod.InitContainers {
// 			err = dotGraph.AddNode("G", GetPrettyString(pod.Name+initContainer), map[string]string{"label": "\"initcontainer: " + initContainer + "\""})
// 			if err != nil {
// 				return "", err
// 			}

// 			dotGraph.AddEdge(GetPrettyString(pod.Name), GetPrettyString(pod.Name+initContainer), true, nil)
// 		}
// 	}

// 	// Add service nodes
// 	for _, service := range graph.Services {
// 		err := dotGraph.AddNode("G", GetPrettyString(service.Name), map[string]string{"label": "\"service: " + service.Name + "\""})
// 		if err != nil {
// 			return "", err
// 		}

// 		for _, pod := range graph.Pods {
// 			addService := true
// 			for key, value := range service.Selector {
// 				if _, ok := pod.Labels[key]; !ok {
// 					addService = false
// 					break
// 				}

// 				if value != pod.Labels[key] {
// 					addService = false
// 					break
// 				}
// 			}

// 			// If service selector is nil, it does not have pods
// 			if addService && service.Selector != nil {
// 				err := dotGraph.AddEdge(GetPrettyString(service.Name), GetPrettyString(pod.Name), true, nil)
// 				if err != nil {
// 					return "", err
// 				}
// 			}
// 		}
// 	}

// 	// Add ingress nodes
// 	for _, ingress := range graph.Ingresses {
// 		err := dotGraph.AddNode("G", GetPrettyString(ingress.Name), map[string]string{"label": "\"ingress: " + ingress.Name + "\""})
// 		if err != nil {
// 			return "", err
// 		}

// 		for _, service := range graph.Services {
// 			if ingress.Service == service.Name {
// 				err := dotGraph.AddEdge(GetPrettyString(ingress.Name), GetPrettyString(service.Name), true, nil)
// 				if err != nil {
// 					return "", err
// 				}
// 			}
// 		}
// 	}

// 	return dotGraph.String(), nil
// }
