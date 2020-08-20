package graph

import (
	"kube-graph/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// PodBuilder holds the client to access the k8s api
type PodBuilder struct {
	Client *client.Client
}

// Pod holds the pod information for the graph
type Pod struct {
	Name            string
	Labels          map[string]string
	InitContainers  []string
	Containers      []string
	OwnerReferences []map[string]string
}

// NewPodBuilder returns a new PodBuilder struct
func NewPodBuilder(client *client.Client) *PodBuilder {
	return &PodBuilder{
		Client: client,
	}
}

// GetPod returns the information of a pod that matches the given name
func (p *PodBuilder) GetPod(name string) (Pod, error) {
	pod, err := p.Client.GetPod(name)
	if err != nil {
		return Pod{}, err
	}

	initContainers := []string{}
	for _, initContainer := range pod.Spec.InitContainers {
		initContainers = append(initContainers, initContainer.Name)
	}

	containers := []string{}
	for _, container := range pod.Spec.Containers {
		containers = append(containers, container.Name)
	}

	ownerReferences := []map[string]string{}
	for _, ownerReference := range pod.ObjectMeta.OwnerReferences {
		ownerReferences = append(ownerReferences,
			map[string]string{
				"kind": ownerReference.Kind,
				"name": ownerReference.Name,
			},
		)
	}

	return Pod{
		Name:            pod.Name,
		Labels:          pod.ObjectMeta.Labels,
		InitContainers:  initContainers,
		Containers:      containers,
		OwnerReferences: ownerReferences,
	}, nil
}

// GetPods returns the information of a list of pods that match the given service selector map
func (p *PodBuilder) GetPods(selector map[string]string) ([]Pod, error) {
	if selector == nil {
		return []Pod{}, nil
	}

	pods, err := p.Client.GetPods(metav1.ListOptions{LabelSelector: labels.Set(selector).String()})
	if err != nil {
		return []Pod{}, err
	}

	relatedPods := []Pod{}

	for _, pod := range pods.Items {
		initContainers := []string{}
		for _, initContainer := range pod.Spec.InitContainers {
			initContainers = append(initContainers, initContainer.Name)
		}

		containers := []string{}
		for _, container := range pod.Spec.Containers {
			containers = append(containers, container.Name)
		}

		ownerReferences := []map[string]string{}
		for _, ownerReference := range pod.ObjectMeta.OwnerReferences {
			ownerReferences = append(ownerReferences,
				map[string]string{
					"kind": ownerReference.Kind,
					"name": ownerReference.Name,
				},
			)
		}

		relatedPods = append(relatedPods, Pod{
			Name:            pod.Name,
			Labels:          pod.ObjectMeta.Labels,
			InitContainers:  initContainers,
			Containers:      containers,
			OwnerReferences: ownerReferences,
		})
	}

	return relatedPods, nil
}
