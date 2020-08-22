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
	Graph     *Graph
}

// Graph holds the k8s objects for the graph
type Graph struct {
	Objs []unstructured.Unstructured
}

// NewBuilder returns a new object struct
func NewBuilder(client dynamic.Interface, namespace, kind, name string) *Builder {
	return &Builder{
		Client:    client,
		Namespace: namespace,
		Kind:      kind,
		Name:      name,
		Graph:     &Graph{},
	}
}

// Build gets all the information required to build the graph
func (b *Builder) Build() error {
	klog.V(1).Infoln("build the graph")

	err := GetObject(b.Client, b.Graph, b.Namespace, b.Kind, b.Name)
	if err != nil {
		return err
	}

	err = GetRelatedObjects(b.Client, b.Graph, b.Namespace, b.Kind, b.Name)
	if err != nil {
		return err
	}

	fmt.Println(b.Graph.Objs)

	// dotg, err := BuildDOTGraph(b.Graph)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(dotg)

	return nil
}

// GetObject gets the searched object
func GetObject(client dynamic.Interface, graph *Graph, namespace, kind, name string) error {
	klog.V(1).Infoln("get initial object")

	gvr, err := GetGroupVersionResource(kind)
	if err != nil {
		return err
	}

	var ri dynamic.ResourceInterface
	ri = client.Resource(gvr).Namespace(namespace)

	obj, err := ri.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	graph.Objs = append(graph.Objs, *obj)

	return nil
}

// GetRelatedObjects gets the related objects
func GetRelatedObjects(client dynamic.Interface, graph *Graph, namespace, kind, name string) error {
	klog.V(1).Infoln("get information of the objects related to the initial object")

	gvrList := GetAllGroupVersionResourcesExcept(kind)

	for _, gvr := range gvrList {
		var ri dynamic.ResourceInterface

		ri = client.Resource(gvr).Namespace(namespace)

		objList, err := ri.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, obj := range objList.Items {
			graph.Objs = append(graph.Objs, obj)
		}
	}

	return nil
}

func GetGroupVersionResource(kind string) (schema.GroupVersionResource, error) {
	if gvr, ok := GetSupportedGroupVersionResources()[kind]; ok {
		return gvr, nil
	}

	return schema.GroupVersionResource{}, fmt.Errorf("kind '%s' not found in supported GroupVersionResources", kind)
}

func GetAllGroupVersionResourcesExcept(kind string) []schema.GroupVersionResource {
	gvrList := []schema.GroupVersionResource{}

	for key, value := range GetSupportedGroupVersionResources() {
		if key != kind {
			gvrList = append(gvrList, value)
		}
	}

	return gvrList
}

func GetAllGroupVersionResources() []schema.GroupVersionResource {
	gvrList := []schema.GroupVersionResource{}

	for _, value := range GetSupportedGroupVersionResources() {
		gvrList = append(gvrList, value)
	}

	return gvrList
}

func GetSupportedGroupVersionResources() map[string]schema.GroupVersionResource {
	return map[string]schema.GroupVersionResource{
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
	}
}

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

// GetPrettyString returns a pretty string that can be used as a dot node name
func GetPrettyString(ugly string) string {
	return strings.ReplaceAll(ugly, "-", "")
}
