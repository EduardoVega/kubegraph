package graph

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type MockClient struct{}

type MockNamespaceableResourceInterface struct{}

type MockResourceInterface struct {
	Resource string
}

func (c MockClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return MockResourceInterface{
		resource.Resource,
	}
}

func (r MockResourceInterface) Namespace(string) dynamic.ResourceInterface {
	return MockResourceInterface{
		r.Resource,
	}
}

func (r MockResourceInterface) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (r MockResourceInterface) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (r MockResourceInterface) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return nil, nil
}

func (r MockResourceInterface) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return nil
}

func (r MockResourceInterface) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (r MockResourceInterface) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	u := unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind": "Service",
			"metadata": map[string]interface{}{
				"name": "service-foo",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app":     "foo",
					"version": "v1",
				},
			},
		},
	}

	return &u, nil
}

func (r MockResourceInterface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	switch r.Resource {
	case "pods":
		return &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-foo-1",
							"labels": map[string]interface{}{
								"app":     "foo",
								"version": "v1",
							},
							"ownerReferences": []interface{}{
								map[string]interface{}{
									"uid": "1d1fcfc1-6f23-4578-9b70-8361a733ab26",
								},
							},
						},
					},
				},
				unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Pod",
						"metadata": map[string]interface{}{
							"name": "pod-foo-2",
							"labels": map[string]interface{}{
								"app":     "foo",
								"version": "v2",
							},
							"ownerReferences": []interface{}{
								map[string]interface{}{
									"uid": "1d1fcfc1-6f23-4578-9b70-8361a733ab26",
								},
							},
						},
					},
				},
			},
		}, nil

	case "services":
		return &unstructured.UnstructuredList{}, nil
	case "ingresses":
		return &unstructured.UnstructuredList{}, nil
	case "replicasets":
		return &unstructured.UnstructuredList{}, nil
	case "deployments", "daemonsets":
		return &unstructured.UnstructuredList{}, nil
	case "statefulsets":
		return &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Statefulset",
						"metadata": map[string]interface{}{
							"name": "statefulset-foo-1",
							"uid":  "1d1fcfc1-6f23-4578-9b70-8361a733ab26",
						},
					},
				},
				unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Statefulset",
						"metadata": map[string]interface{}{
							"name": "statefulset-foo-2",
							"uid":  "1d1fcfc1-6f23-4578-9b70-8361a733ab20",
						},
					},
				},
			},
		}, nil
	}

	return nil, nil
}

func (r MockResourceInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (r MockResourceInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, nil
}

func TestBuild(t *testing.T) {
	o := &bytes.Buffer{}

	c := MockClient{}

	b := NewBuilder(c, o, false, "default", "service", "service-foo")

	b.Build()

	expected := "\n[Service] service-foo\n\t\t┌── [Statefulset] statefulset-foo-1\n\t└── [Pod] pod-foo-1\n\n"

	if o.String() != expected {
		t.Errorf("Returned graph was incorrect,\ngot:\n%s\nwant:\n%s", o.String(), expected)
	}
}

func TestGetRelatedKinds(t *testing.T) {

	tests := []struct {
		Kind     string
		Expected map[string][]string
	}{
		{
			"pod",
			map[string][]string{
				"upper": []string{"service", "replicaset", "statefulset"},
				"lower": []string{},
			},
		},
		{
			"service",
			map[string][]string{
				"upper": []string{"ingress"},
				"lower": []string{"pod"},
			},
		},
		{
			"ingress",
			map[string][]string{
				"upper": []string{},
				"lower": []string{"service"},
			},
		},
		{
			"replicaset",
			map[string][]string{
				"upper": []string{"deployment", "daemonset"},
				"lower": []string{"pod"},
			},
		},
		{
			"daemonset",
			map[string][]string{
				"upper": []string{},
				"lower": []string{"replicaset"},
			},
		},
		{
			"statefulset",
			map[string][]string{
				"upper": []string{},
				"lower": []string{"pod"},
			},
		},
	}

	for _, test := range tests {
		r := GetRelatedKinds(test.Kind)

		if !reflect.DeepEqual(r["upper"], test.Expected["upper"]) || !reflect.DeepEqual(r["lower"], test.Expected["lower"]) {
			t.Errorf("Returned result was incorrect, got: %v want: %v", r, test.Expected)
		}

	}
}

func TestGetGroupVersionResource(t *testing.T) {

	tests := []struct {
		Kind     string
		Expected string
	}{
		{
			"po",
			"pods",
		},
		{
			"svc",
			"services",
		},
		{
			"ing",
			"ingresses",
		},
		{
			"rs",
			"replicasets",
		},
		{
			"deploy",
			"deployments",
		},
		{
			"ds",
			"daemonsets",
		},
		{
			"sts",
			"statefulsets",
		},
		{
			"foo",
			"kind 'foo' not supported",
		},
	}

	for _, test := range tests {
		r, err := GetGroupVersionResource(test.Kind)

		if err != nil {
			if err.Error() != fmt.Sprintf("kind '%s' not supported", test.Kind) {
				t.Errorf("Returned result was incorrect, got: %s want: %s", err.Error(), test.Expected)
			}
		} else {
			if r.Resource != test.Expected {
				t.Errorf("Returned result was incorrect, got: %s want: %s", r.Resource, test.Expected)
			}
		}
	}
}
