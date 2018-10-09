package gorgonia

import (
	"fmt"
)

//data for Built
type data struct {
	fun        interface{}
	nodes      []*Node
	retSame    bool
	axis       int
	p          int
	nums       []int
	nodeBefore bool
	name       string
	gName      string
}

func (d *data) setFun(fun interface{}) *data {
	d.fun = fun
	return d
}

func (d *data) setNodes(nodes []*Node) *data {
	d.nodes = nodes
	return d
}

func (d *data) setRetSame(retSame bool) *data {
	d.retSame = retSame
	return d
}

func (d *data) setAxis(axis int) *data {
	d.axis = axis
	return d
}

func (d *data) setP(p int) *data {
	d.p = p
	return d
}

func (d *data) setNums(nums []int) *data {
	d.nums = nums
	return d
}

func (d *data) setNodeBefore(b bool) *data {
	d.nodeBefore = b
	return d
}

func (d *data) setName(name string) *data {
	d.name = name
	return d
}

func (d *data) setGroupName(name string) *data {
	d.gName = name
	return d
}

//Built takes a gorgonia functions and list of node, and return result node
func Built(fun interface{}, nodes ...*Node) (rtn *Node) {
	d := &data{fun: fun, nodes: nodes}

	rtn, _ = resolve(d)
	return
}

//BuiltGrad takes list of node, first is scalar cost node and the remaining nodes is regards to input nodes, and returns the gradient
func BuiltGrad(nodes ...*Node) (ns Nodes) {
	d := &data{fun: Grad, nodes: nodes}
	_, ns = resolve(d)
	return
}

//Built takes a gorgonia functions and list of node, and return result node
func (n *Node) Built(fun interface{}, nodes ...*Node) (rtn *Node) {
	d := &data{fun: fun, nodes: nodes}
	if n != nil {
		d.nodes = append([]*Node{n}, d.nodes...)
	}

	rtn, _ = resolve(d)
	return
}

//BuiltGrad takes list of node, first is scalar cost node and the remaining nodes is regards to input nodes, and returns the gradient
func (n *Node) BuiltGrad(nodes ...*Node) (ns Nodes) {
	d := &data{fun: Grad, nodes: nodes}
	if n != nil {
		d.nodes = append([]*Node{n}, d.nodes...)
	}

	_, ns = resolve(d)
	return
}

//SetName set name for Node
func (n *Node) SetName(name string) *Node {
	n.name = name
	return n
}

//SetGroupName set group name for Node
func (n *Node) SetGroupName(name string) *Node {
	n.group = name
	return n
}

//resolve call the gorgonia functions
func resolve(d *data) (n *Node, ns Nodes) {
	var err error

	fun := d.fun
	nodes := d.nodes
	retSame := d.retSame
	axis := d.axis
	p := d.p
	nums := d.nums
	name := d.name
	gName := d.gName

	switch f := fun.(type) {
	case func(*Node) (*Node, error):
		n, err = f(nodes[0])
	case func(*Node, *Node) (*Node, error):
		n, err = f(nodes[0], nodes[1])
	case func(*Node, *Node, bool) (*Node, error):
		n, err = f(nodes[0], nodes[1], retSame)
	case func(*Node, int) (*Node, error):
		n, err = f(nodes[0], axis)
	case func(*Node, int, int) (*Node, error):
		n, err = f(nodes[0], axis, p)
	case func(*Node, ...int) (*Node, error):
		n, err = f(nodes[0], nums...) //TODO issue FIXME
	case func(*Node, ...NodeConsOpt) (*Node, error):
		n, err = f(nodes[0])
	case func(int, ...*Node) (*Node, error):
		n, err = f(axis, nodes...)
	case func(...*Node) (*Node, error):
		n, err = f(nodes...)
	case func(*Node, ...*Node) (Nodes, error):
		ns, err = f(nodes[0], nodes[1:]...)
		if len(ns) > 0 {
			n = ns[0]
		}
	default:
		fmt.Printf("I don't know about type %T!\n", f)
	}

	if err != nil || n == nil {
		panic(err)
	}

	if name != "" {
		WithName(name)(n)
	}
	if gName != "" {
		WithGroupName(gName)(n)
	}

	return
}
