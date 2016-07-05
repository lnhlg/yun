package yun

type (
	node struct {
		path 	string
		ntype	nodeType
		length	int
		next 	*node
	}
)

func addNode(nod *node, path string, nodType nodeType) *node {
	if nod == nil {
		nod = new(node)
	} else {
		nod.next = new(node)
		nod = nod.next
	}
	nod.path = path
	nod.length = len(path)
	nod.ntype = nodType

	return nod
}

