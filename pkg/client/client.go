package client

import (
	"context"

	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
)

// Client defines attributes required to call k8s objects
type Client struct {
	Clientset *kubernetes.Clientset
	Namespace string
}

// NewClient returns a new Client struct
func NewClient(clientset *kubernetes.Clientset, namespace string) *Client {
	return &Client{
		Clientset: clientset,
		Namespace: namespace,
	}
}

// GetPod returns a pod
func (c *Client) GetPod(name string) (pod *v1.Pod, err error) {
	pod, err = c.Clientset.CoreV1().Pods(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return
}

// GetPods returns a list of pods
func (c *Client) GetPods(options metav1.ListOptions) (pods *v1.PodList, err error) {
	pods, err = c.Clientset.CoreV1().Pods(c.Namespace).List(context.TODO(), options)
	return
}

// GetService returns a service
func (c *Client) GetService(name string) (service *v1.Service, err error) {
	service, err = c.Clientset.CoreV1().Services(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return
}

// GetServices returns a list of services
func (c *Client) GetServices(options metav1.ListOptions) (services *v1.ServiceList, err error) {
	services, err = c.Clientset.CoreV1().Services(c.Namespace).List(context.TODO(), options)
	return
}

// GetIngress returns an ingress
func (c *Client) GetIngress(name string) (ingress *netv1beta1.Ingress, err error) {
	ingress, err = c.Clientset.NetworkingV1beta1().Ingresses(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return
}

// GetIngresses returns a list of ingresses
func (c *Client) GetIngresses(options metav1.ListOptions) (ingresses *netv1beta1.IngressList, err error) {
	ingresses, err = c.Clientset.NetworkingV1beta1().Ingresses(c.Namespace).List(context.TODO(), options)
	return
}
