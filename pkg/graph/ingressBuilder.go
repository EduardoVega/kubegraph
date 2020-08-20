package graph

import (
	"kube-graph/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressBuilder struct {
	Client *client.Client
}

type Ingress struct {
	Name    string
	Host    string
	Path    string
	Service string
}

func NewIngressBuilder(client *client.Client) *IngressBuilder {
	return &IngressBuilder{
		Client: client,
	}
}

func (i *IngressBuilder) GetIngress(name string) ([]Ingress, error) {
	ingress, err := i.Client.GetIngress(name)
	if err != nil {
		return []Ingress{}, err
	}

	ingresses := []Ingress{}

	for _, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			ingresses = append(ingresses, Ingress{
				Name:    ingress.Name,
				Host:    rule.Host,
				Path:    path.Path,
				Service: path.Backend.ServiceName,
			})
		}
	}

	return ingresses, nil
}

func (i *IngressBuilder) GetIngresses(services []Service, options metav1.ListOptions) ([]Ingress, error) {
	ingresses, err := i.Client.GetIngresses(options)
	if err != nil {
		return []Ingress{}, err
	}

	relatedIngresses := []Ingress{}

	for _, ingress := range ingresses.Items {
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				for _, service := range services {
					if path.Backend.ServiceName == service.Name {
						relatedIngresses = append(relatedIngresses, Ingress{
							Name:    ingress.Name,
							Host:    rule.Host,
							Path:    path.Path,
							Service: path.Backend.ServiceName,
						})
					}
				}
			}
		}
	}

	return relatedIngresses, nil
}
