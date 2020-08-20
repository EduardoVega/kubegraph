package graph

import (
	"fmt"
	"kube-graph/pkg/client"
	"strings"

	"github.com/awalterschulze/gographviz"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder holds the information to build the graph
type Builder struct {
	Client *client.Client
	Kind   string
	Name   string
	Graph  *Graph
}

// Graph holds the k8s objects for the graph
type Graph struct {
	Pods      []Pod
	Services  []Service
	Ingresses []Ingress
}

// NewBuilder returns a new object struct
func NewBuilder(client *client.Client, kind string, name string) *Builder {
	return &Builder{
		Client: client,
		Kind:   kind,
		Name:   name,
		Graph:  &Graph{},
	}
}

// Build gets all the information required to build the graph
func (b *Builder) Build() error {
	// Get initial object
	err := GetInitialObject(b.Client, b.Graph, b.Kind, b.Name)
	if err != nil {
		return err
	}

	// Get related objects
	err = GetRelatedObjects(b.Client, b.Graph, b.Kind)
	if err != nil {
		return err
	}

	dotg, err := BuildDOTGraph(b.Graph)
	if err != nil {
		return err
	}

	fmt.Println(dotg)

	return nil
}

// GetInitialObject gets the initial object information for the graph
func GetInitialObject(client *client.Client, graph *Graph, kind, name string) error {
	switch kind {
	case "pod":
		podBuilder := NewPodBuilder(client)
		pod, err := podBuilder.GetPod(name)
		if err != nil {
			return err
		}

		graph.Pods = []Pod{pod}

	case "service":
		serviceBuilder := NewServiceBuilder(client)
		service, err := serviceBuilder.GetService(name)
		if err != nil {
			return err
		}

		graph.Services = []Service{service}

	case "ingress":
		ingressBuilder := NewIngressBuilder(client)
		ingresses, err := ingressBuilder.GetIngress(name)
		if err != nil {
			return err
		}

		graph.Ingresses = ingresses

	default:
		return fmt.Errorf("TYPE not supported. Run: kubectl graph -h")

	}

	return nil
}

// GetRelatedObjects gets the related objects information for the graph
func GetRelatedObjects(client *client.Client, graph *Graph, kind string) error {
	switch kind {
	case "pod":
		serviceBuilder := NewServiceBuilder(client)
		services, err := serviceBuilder.GetServices(graph.Pods[0].Labels, metav1.ListOptions{})
		if err != nil {
			return err
		}

		graph.Services = services

		ingressBuilder := NewIngressBuilder(client)
		ingresses, err := ingressBuilder.GetIngresses(graph.Services, metav1.ListOptions{})
		if err != nil {
			return err
		}

		graph.Ingresses = ingresses

	case "service":
		podBuilder := NewPodBuilder(client)
		pods, err := podBuilder.GetPods(graph.Services[0].Selector)
		if err != nil {
			return err
		}

		graph.Pods = pods

		ingressBuilder := NewIngressBuilder(client)
		ingresses, err := ingressBuilder.GetIngresses(graph.Services, metav1.ListOptions{})
		if err != nil {
			return err
		}

		graph.Ingresses = ingresses

	case "ingress":
		for _, ingress := range graph.Ingresses {
			serviceBuilder := NewServiceBuilder(client)
			service, err := serviceBuilder.GetService(ingress.Service)
			if err != nil {
				if !strings.Contains(err.Error(), "not found") {
					return err
				}
			} else {
				graph.Services = append(graph.Services, service)
			}
		}

		graph.Pods = []Pod{}
		for _, service := range graph.Services {
			podBuilder := NewPodBuilder(client)
			pods, err := podBuilder.GetPods(service.Selector)
			if err != nil {
				return err
			}

			for _, pod := range pods {
				graph.Pods = append(graph.Pods, pod)
			}
		}
	}

	return nil
}

// BuildDOTGraph returns a DOT graph populated with obtained k8s objects
func BuildDOTGraph(graph *Graph) (string, error) {
	dotGraph := gographviz.NewGraph()

	err := dotGraph.SetName("G")
	if err != nil {
		return "", err
	}

	err = dotGraph.SetDir(true)
	if err != nil {
		return "", err
	}

	// Add pod nodes
	for _, pod := range graph.Pods {

		err = dotGraph.AddNode("G", GetPrettyString(pod.Name), map[string]string{"label": "\"pod: " + pod.Name + "\""})
		if err != nil {
			return "", err
		}

		for _, container := range pod.Containers {
			err = dotGraph.AddNode("G", GetPrettyString(pod.Name+container), map[string]string{"label": "\"container: " + container + "\""})
			if err != nil {
				return "", err
			}

			dotGraph.AddEdge(GetPrettyString(pod.Name), GetPrettyString(pod.Name+container), true, nil)
		}

		for _, initContainer := range pod.InitContainers {
			err = dotGraph.AddNode("G", GetPrettyString(pod.Name+initContainer), map[string]string{"label": "\"initcontainer: " + initContainer + "\""})
			if err != nil {
				return "", err
			}

			dotGraph.AddEdge(GetPrettyString(pod.Name), GetPrettyString(pod.Name+initContainer), true, nil)
		}
	}

	// Add service nodes
	for _, service := range graph.Services {
		err := dotGraph.AddNode("G", GetPrettyString(service.Name), map[string]string{"label": "\"service: " + service.Name + "\""})
		if err != nil {
			return "", err
		}

		for _, pod := range graph.Pods {
			addService := true
			for key, value := range service.Selector {
				if _, ok := pod.Labels[key]; !ok {
					addService = false
					break
				}

				if value != pod.Labels[key] {
					addService = false
					break
				}
			}

			// If service selector is nil, it does not have pods
			if addService && service.Selector != nil {
				err := dotGraph.AddEdge(GetPrettyString(service.Name), GetPrettyString(pod.Name), true, nil)
				if err != nil {
					return "", err
				}
			}
		}
	}

	// Add ingress nodes
	for _, ingress := range graph.Ingresses {
		err := dotGraph.AddNode("G", GetPrettyString(ingress.Name), map[string]string{"label": "\"ingress: " + ingress.Name + "\""})
		if err != nil {
			return "", err
		}

		for _, service := range graph.Services {
			if ingress.Service == service.Name {
				err := dotGraph.AddEdge(GetPrettyString(ingress.Name), GetPrettyString(service.Name), true, nil)
				if err != nil {
					return "", err
				}
			}
		}
	}

	return dotGraph.String(), nil
}

// GetPrettyString returns a pretty string that can be used as a dot node name
func GetPrettyString(ugly string) string {
	return strings.ReplaceAll(ugly, "-", "")
}
