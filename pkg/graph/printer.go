package graph

import (
	"fmt"
	"io"

	"github.com/awalterschulze/gographviz"
)

type Printer struct {
	ObjData
	Out io.Writer
}

func NewPrinter(objData ObjData, out io.Writer) *Printer {
	return &Printer{
		ObjData: objData,
		Out:     out,
	}
}

func (p *Printer) Print(printerType string) (err error) {
	var g string

	if printerType == "dot" {
		g, err = p.PrintDotGraph()
	} else {
		g, err = p.PrintTreeGraph()
	}

	if err != nil {
		return err
	}

	fmt.Fprint(p.Out, g)

	return
}

func (p *Printer) PrintTreeGraph() (string, error) {
	return "", nil
}

func (p *Printer) PrintDotGraph() (string, error) {
	g := gographviz.NewGraph()

	err := g.SetName("W")
	// Directed graph to show relationships between nodes (->)
	err = g.SetDir(true)
	// Disable multiple edges between nodes
	err = g.SetStrict(true)
	if err != nil {
		return "", err
	}

	_, err = CreateNodesEdges(p.ObjData, g)
	if err != nil {
		return "", err
	}

	return g.String(), nil
}

func CreateNodesEdges(o ObjData, g *gographviz.Graph) (string, error) {
	err := g.AddNode("W", GetPrettyString(o.Obj.GetKind()+o.Obj.GetName()), map[string]string{"label": "\"" + o.Obj.GetKind() + ": " + o.Obj.GetName() + "\""})
	if err != nil {
		return "", err
	}

	for _, r := range o.RelatedObjsData {
		n, err := CreateNodesEdges(r, g)
		if err != nil {
			return "", err
		}

		if r.Hierarchy == "upper" {
			g.AddEdge(n, GetPrettyString(o.Obj.GetKind()+o.Obj.GetName()), true, nil)
		}

		if r.Hierarchy == "lower" {
			g.AddEdge(GetPrettyString(o.Obj.GetKind()+o.Obj.GetName()), n, true, nil)
		}

	}

	return GetPrettyString(o.Obj.GetKind() + o.Obj.GetName()), nil
}
