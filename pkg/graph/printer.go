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
	g := ""

	if p.DotGraph {
		gv := gographviz.NewGraph()

		err := gv.SetName("W")
		if err != nil {
			return err
		}
		// Directed graph to show relationships between nodes (->)
		err = gv.SetDir(true)
		if err != nil {
			return err
		}
		// Disable multiple edges between nodes
		err = gv.SetStrict(true)
		if err != nil {
			return err
		}

		_, err = CreateDotGraph(p.ObjData, gv)
		if err != nil {
			return err
		}

		g = gv.String()
	} else {
		g = fmt.Sprintf("\n%s\n\n", CreateTreeGraph(p.ObjData, "", ""))
	}

	fmt.Fprint(p.Out, g)

	return
}

// CreateTreeGraph returns a string holding the tree graph
func CreateTreeGraph(o ObjData, graph, format string) string {

	if graph == "" {
		graph = fmt.Sprintf("[%s] %s", o.Obj.GetKind(), o.Obj.GetName())
	} else if o.Hierarchy == "upper" {
		graph = fmt.Sprintf("%s┌── [%s] %s", format, o.Obj.GetKind(), o.Obj.GetName())
	} else if o.Hierarchy == "lower" {
		graph = fmt.Sprintf("%s└── [%s] %s", format, o.Obj.GetKind(), o.Obj.GetName())
	}

	format = format + "\t"

	for _, r := range o.RelatedObjsData {
		relatedGraph := CreateTreeGraph(r, graph, format)

		if r.Hierarchy == "upper" {
			graph = relatedGraph + "\n" + graph
		} else {
			graph = graph + "\n" + relatedGraph
		}
	}

	return graph
}

// CreateDotGraph returns a string holding the dot graph
func CreateDotGraph(o ObjData, g *gographviz.Graph) (string, error) {
	err := g.AddNode("W", GetPrettyString(o.Obj.GetKind()+o.Obj.GetName()), map[string]string{"label": "\"" + o.Obj.GetKind() + ": " + o.Obj.GetName() + "\""})
	if err != nil {
		return "", err
	}

	for _, r := range o.RelatedObjsData {
		n, err := CreateDotGraph(r, g)
		if err != nil {
			return "", err
		}

		if r.Hierarchy == "upper" {
			err := g.AddEdge(n, GetPrettyString(o.Obj.GetKind()+o.Obj.GetName()), true, nil)
			if err != nil {
				return "", err
			}
		}

		if r.Hierarchy == "lower" {
			err := g.AddEdge(GetPrettyString(o.Obj.GetKind()+o.Obj.GetName()), n, true, nil)
			if err != nil {
				return "", err
			}
		}

	}

	return GetPrettyString(o.Obj.GetKind() + o.Obj.GetName()), nil
}
