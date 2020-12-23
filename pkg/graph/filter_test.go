package graph

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestFilter(t *testing.T) {

	tests := []struct {
		Obj        unstructured.Unstructured
		RelatedObj unstructured.Unstructured
		Expected   bool
	}{
		{
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Statefulset",
					"metadata": map[string]interface{}{
						"uid": "1d1fcfc1-6f23-4578-9b70-8361a733ab26",
					},
				},
			},
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"uid": "1d1fcfc1-6f23-4578-9b70-8361a733ab26",
							},
						},
					},
				},
			},
			true,
		},
		{
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Replicaset",
					"metadata": map[string]interface{}{
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
					"kind": "Deployment",
					"metadata": map[string]interface{}{
						"uid": "1d1fcfc1-6f23-4578-9b70-8361a733ab26",
					},
				},
			},
			true,
		},
		{
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Ingress",
					"spec": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"backend": map[string]interface{}{
												"serviceName": "service-foo",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Service",
					"metadata": map[string]interface{}{
						"name": "service-foo",
					},
				},
			},
			true,
		},
		{
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Service",
					"metadata": map[string]interface{}{
						"name": "service-foo",
					},
				},
			},
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Ingress",
					"spec": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"backend": map[string]interface{}{
												"serviceName": "service-bar",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app":     "foo",
							"version": "v1",
						},
					},
				},
			},
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Service",
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"app":     "foo",
							"version": "v1",
						},
					},
				},
			},
			true,
		},
		{
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Service",
					"spec": map[string]interface{}{
						"selector": map[string]interface{}{
							"app":     "foo",
							"version": "v2",
						},
					},
				},
			},
			unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app":     "foo",
							"version": "v1",
						},
					},
				},
			},
			false,
		},
	}

	for _, test := range tests {
		f := NewFilter()
		r := f.FilterObj(test.Obj, test.RelatedObj)

		if r != test.Expected {
			t.Errorf("Returned result was incorrect, got: %t want: %t", r, test.Expected)
		}
	}
}
