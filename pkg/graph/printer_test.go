package graph

import (
	"bytes"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestPrint(t *testing.T) {

	tests := []struct {
		ObjData           ObjData
		ExpectedTreeGraph string
		ExpectedDotGraph  string
	}{
		{
			ObjData{
				unstructured.Unstructured{
					Object: map[string]interface{}{
						"kind": "Service",
						"metadata": map[string]interface{}{
							"name": "service-foo",
						},
					},
				},
				"",
				[]ObjData{
					{
						unstructured.Unstructured{
							Object: map[string]interface{}{
								"kind": "Ingress",
								"metadata": map[string]interface{}{
									"name": "ingress-foo",
								},
							},
						},
						"upper",
						[]ObjData{},
					},
					{
						unstructured.Unstructured{
							Object: map[string]interface{}{
								"kind": "Pod",
								"metadata": map[string]interface{}{
									"name": "pod-foo",
								},
							},
						},
						"lower",
						[]ObjData{},
					},
				},
			},
			"\n\t┌── [Ingress] ingress-foo\n[Service] service-foo\n\t└── [Pod] pod-foo\n\n",
			`strict digraph W {
	Ingressingressfoo->Serviceservicefoo;
	Serviceservicefoo->Podpodfoo;
	Ingressingressfoo [ label="Ingress: ingress-foo" ];
	Podpodfoo [ label="Pod: pod-foo" ];
	Serviceservicefoo [ label="Service: service-foo" ];

}
`,
		},
	}

	for _, test := range tests {
		resultTreeGraph := &bytes.Buffer{}
		p1 := NewPrinter(test.ObjData, false, resultTreeGraph)
		p1.Print()

		if resultTreeGraph.String() != test.ExpectedTreeGraph {
			t.Errorf("Returned graph was incorrect,\ngot:\n%s\nwant:\n%s", resultTreeGraph.String(), test.ExpectedTreeGraph)
		}

		resultDotGraph := &bytes.Buffer{}
		p2 := NewPrinter(test.ObjData, true, resultDotGraph)
		p2.Print()

		if resultDotGraph.String() != test.ExpectedDotGraph {
			t.Errorf("Returned graph was incorrect,\ngot:\n%s\nwant:\n%s", resultDotGraph.String(), test.ExpectedDotGraph)
		}
	}
}
