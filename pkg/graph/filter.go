package graph

import (
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

// Filter holds the functions to filter the related objects
type Filter struct{}

// NewFilter returns a new Filter struct
func NewFilter() *Filter {
	return &Filter{}
}

// FilterObj filters the related objects based on the obj and related obj kind
func (f *Filter) FilterObj(obj unstructured.Unstructured, relatedObj unstructured.Unstructured) bool {
	switch strings.ToLower(obj.GetKind()) {
	case "pod":
		switch strings.ToLower(relatedObj.GetKind()) {
		case "service":
			return FilterByLabelSelector(GetSelector(relatedObj), obj.GetLabels())
		case "replicaset", "statefulset", "daemonset":
			return FilterByOwnerReferenceUID(obj.GetOwnerReferences(), relatedObj.GetUID())
		}
	case "service":
		switch strings.ToLower(relatedObj.GetKind()) {
		case "ingress":
			return FilterByServiceName(obj.GetName(), GetBackendNames(relatedObj))
		case "pod":
			return FilterByLabelSelector(GetSelector(obj), relatedObj.GetLabels())
		}
	case "ingress":
		switch strings.ToLower(relatedObj.GetKind()) {
		case "service":
			return FilterByServiceName(relatedObj.GetName(), GetBackendNames(obj))
		}
	case "replicaset":
		switch strings.ToLower(relatedObj.GetKind()) {
		case "deployment":
			return FilterByOwnerReferenceUID(obj.GetOwnerReferences(), relatedObj.GetUID())
		case "pod":
			return FilterByOwnerReferenceUID(relatedObj.GetOwnerReferences(), obj.GetUID())

		}
	case "deployment":
		switch strings.ToLower(relatedObj.GetKind()) {
		case "replicaset":
			return FilterByOwnerReferenceUID(relatedObj.GetOwnerReferences(), obj.GetUID())
		}
	case "statefulset", "daemonset":
		switch strings.ToLower(relatedObj.GetKind()) {
		case "pod":
			return FilterByOwnerReferenceUID(relatedObj.GetOwnerReferences(), obj.GetUID())
		}

	}

	return false
}

// GetBackendNames returns the backend services configured in an ingress
func GetBackendNames(ingressObj unstructured.Unstructured) []string {
	backendServiceNames := []string{}
	rules := ingressObj.Object["spec"].(map[string]interface{})["rules"]

	if rules != nil {
		for _, r := range rules.([]interface{}) {
			http := r.(map[string]interface{})["http"]
			if http != nil {
				paths := http.(map[string]interface{})["paths"]
				if paths != nil {
					for _, p := range paths.([]interface{}) {
						backend := p.(map[string]interface{})["backend"]
						backendServiceNames = append(backendServiceNames, backend.(map[string]interface{})["serviceName"].(string))
					}
				}
			}
		}
	}

	return backendServiceNames
}

// GetSelector returns the label selector from a service
func GetSelector(serviceObj unstructured.Unstructured) map[string]interface{} {
	s := serviceObj.Object["spec"].(map[string]interface{})["selector"]

	if s != nil {
		return s.(map[string]interface{})
	}

	return nil
}

// FilterByServiceName returns true if a service name is found in
// an ingress backend service name list
func FilterByServiceName(serviceName string, backendNames []string) bool {
	for _, b := range backendNames {
		if serviceName == b {
			return true
		}
	}

	return false
}

// FilterByOwnerReferenceUID returns true if an object UID is found
// in an owner reference list
func FilterByOwnerReferenceUID(ownerReferences []v1.OwnerReference, relatedObjUID types.UID) bool {
	for _, r := range ownerReferences {
		if r.UID == relatedObjUID {
			return true
		}
	}

	return false
}

// FilterByLabelSelector returns true if selector keys and values are found in a labels map
func FilterByLabelSelector(selector map[string]interface{}, labels map[string]string) bool {
	if selector == nil {
		return false
	}

	for key, value := range selector {
		if _, ok := labels[key]; !ok {
			return false
		}

		if value != labels[key] {
			return false
		}
	}

	return true
}
