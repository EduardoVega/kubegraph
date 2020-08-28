package graph

import (
	"fmt"
	"io"

	"github.com/awalterschulze/gographviz"
)

// Printer holds the object data and graph type to print
type Printer struct {
	ObjData
	DotGraph bool
	Out      io.Writer
}

// NewPrinter returns a new Printer struct
func NewPrinter(objData ObjData, dotGraph bool, out io.Writer) *Printer {
	return &Printer{
		ObjData:  objData,
		DotGraph: dotGraph,
		Out:      out,
	}
}

// Print prints the tree or dot graph
func (p *Printer) Print() (err error) {
	var g string

	if p.DotGraph {
		g, err = CreateDotGraph(p.ObjData)
	} else {
		g, err = CreateTreeGraph(p.ObjData)
	}

	if err != nil {
		return err
	}

	fmt.Fprint(p.Out, g)

	return
}

// CreateTreeGraph returns a string holding the tree graph
func CreateTreeGraph(o ObjData) (string, error) {
	return "", nil
}

// CreateDotGraph returns a string holding the dot graph
func CreateDotGraph(o ObjData) (string, error) {
	g := gographviz.NewGraph()

	err := g.SetName("W")
	// Directed graph to show relationships between nodes (->)
	err = g.SetDir(true)
	// Disable multiple edges between nodes
	err = g.SetStrict(true)
	if err != nil {
		return "", err
	}

	_, err = CreateNodesEdges(o, g)
	if err != nil {
		return "", err
	}

	return g.String(), nil
}

// CreateNodesEdges creates dot nodes using the objects and dot edges using the relationships between the objects
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
