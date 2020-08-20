package graph

import (
	"kube-graph/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceBuilder holds the client to access the k8s api
type ServiceBuilder struct {
	Client *client.Client
}

// Service holds the pod information for the graph
type Service struct {
	Name         string
	Selector     map[string]string
	ExternalName string
}

// NewServiceBuilder returns a new ServiceBuilder struct
func NewServiceBuilder(client *client.Client) *ServiceBuilder {
	return &ServiceBuilder{
		Client: client,
	}
}

// GetService returns the information of a service that matches the given name
func (s *ServiceBuilder) GetService(name string) (Service, error) {
	service, err := s.Client.GetService(name)
	if err != nil {
		return Service{}, err
	}

	return Service{
		Name:         service.Name,
		Selector:     service.Spec.Selector,
		ExternalName: service.Spec.ExternalName,
	}, nil
}

// GetServices returns the information of a list of services that match the given pod labels map
func (s *ServiceBuilder) GetServices(labels map[string]string, options metav1.ListOptions) ([]Service, error) {
	if labels == nil {
		return []Service{}, nil
	}

	services, err := s.Client.GetServices(options)
	if err != nil {
		return []Service{}, err
	}

	relatedServices := []Service{}

	// Compare selector and labels
	for _, service := range services.Items {
		addService := true

		for key, value := range service.Spec.Selector {
			if _, ok := labels[key]; !ok {
				addService = false
				break
			}

			if value != labels[key] {
				addService = false
				break
			}
		}

		// Add service to the list if selector matches labels
		if addService && service.Spec.Selector != nil {
			relatedServices = append(relatedServices, Service{
				Name:         service.Name,
				Selector:     service.Spec.Selector,
				ExternalName: service.Spec.ExternalName,
			})
		}
	}

	return relatedServices, nil
}
