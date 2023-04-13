// Package csv provides a builder for csv data
package csv

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	headerConjunction      = "."
	DefaultFieldSeparator  = ","
	DefaultRecordSeparator = "\n"
)

// Marshal marshals the object into JSON then converts JSON to CSV then returns the CSV.
func Marshal(input interface{}, fieldSeparator, recordSeparator string) ([]byte, error) {
	type jsonRawOutput = map[string]interface{}
	var outputItemsRaw []jsonRawOutput

	jsonRawInput, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(jsonRawInput, &outputItemsRaw); err != nil {
		return nil, err
	}

	nodes := make([]*node, len(outputItemsRaw))

	for i, tailRaw := range outputItemsRaw {
		newNode := newCSVRoot(tailRaw)
		nodes[i] = &newNode
	}

	for _, node := range nodes {
		if err = node.expandTail(); err != nil {
			return nil, err
		}
	}

	var leaves []*node
	for _, node := range nodes {
		nodeLeaves := node.getLeaves()
		leaves = append(leaves, nodeLeaves...)
	}

	var output strings.Builder
	headers, err := getUniqueHeaders(leaves)
	if err != nil {
		return nil, err
	}
	output.WriteString(strings.Join(headers, fieldSeparator))
	output.WriteString(recordSeparator)
	for _, n := range nodes {
		recordTmp := ""
		for index, header := range headers {
			if index > 0 {
				recordTmp += fieldSeparator
			}
			if n.headToNodeMap[header] != nil {
				recordTmp += fmt.Sprintf("%v", n.headToNodeMap[header].tail)
			}
		}
		output.WriteString(recordTmp)
		output.WriteString(recordSeparator)
	}

	return []byte(output.String()), nil
}

type node struct {
	root          *node
	parent        *node
	headLocal     string
	headFull      string
	tail          interface{}
	children      []*node
	headToNodeMap map[string]*node
}

// newCSVNode returns instance of preconfigured CSV node.
// It contains the information needed to generate a header for data in corresponding CSV field.
func newCSVNode(root, parent *node, headLocal, headFull string, tail interface{}) node {
	return node{
		root:      root,
		parent:    parent,
		headLocal: headLocal,
		headFull:  headFull,
		tail:      tail,
	}
}

// newCSVRoot returns instance of preconfigured CSV node.
// The returned node is a root of the graph.
func newCSVRoot(tail interface{}) node {
	root := node{
		tail:          tail,
		headToNodeMap: make(map[string]*node),
	}
	root.root = &root
	return root
}

// expandTail retrieves the values of all tails in the graph and creates the whole graph structure.
func (n *node) expandTail() error {
	switch nodeTail := n.tail.(type) {
	case string:
		escapedCsvInjectionSigns := escapeCSVInjectionSigns(nodeTail)
		escapedDoubleQuotes := strings.ReplaceAll(escapedCsvInjectionSigns, `"`, `""`)
		enclosedField := fmt.Sprintf("%s%s%s", `"`, escapedDoubleQuotes, `"`)
		n.tail = enclosedField
	case float64, bool:
		n.tail = nodeTail
	case []interface{}:
		for childHeadInt, childTail := range nodeTail {
			childHeadStr := strconv.Itoa(childHeadInt)
			newHeadFull := generateFullHeader(n.headFull, headerConjunction, childHeadStr)
			newNode := newCSVNode(n.root, n, childHeadStr, newHeadFull, childTail)
			n.children = append(n.children, &newNode)
			if err := newNode.expandTail(); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		for childHead, childTail := range nodeTail {
			newHeadFull := generateFullHeader(n.headFull, headerConjunction, childHead)
			newNode := newCSVNode(n.root, n, childHead, newHeadFull, childTail)
			n.children = append(n.children, &newNode)
			if err := newNode.expandTail(); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("error expanding the tail of csv node - type mismatch: %v", n.tail)
	}
	return nil
}

// getLeaves returns all leaves from the subtree starting from the specific node.
func (n *node) getLeaves() []*node {
	var leaves []*node
	if len(n.children) == 0 {
		n.root.headToNodeMap[n.headFull] = n
		return []*node{n}
	}
	for _, child := range n.children {
		childLeaves := child.getLeaves()
		leaves = append(leaves, childLeaves...)
	}
	return leaves
}

// GetHeaders returns an array of headers of all child nodes.
func (n *node) GetHeaders() ([]string, error) {
	var headersSet []string
	if len(n.children) == 0 {
		return []string{n.headFull}, nil
	}
	for _, childNode := range n.children {
		childHeaders, err := childNode.GetHeaders()
		if err != nil {
			return nil, err
		}
		headersSet = append(headersSet, childHeaders...)
	}

	return headersSet, nil
}

func getUniqueHeaders(nodes []*node) ([]string, error) {
	var headersList []string
	for _, n := range nodes {
		nodeHeaders, err := n.GetHeaders()
		if err != nil {
			return nil, err
		}
		headersList = append(headersList, nodeHeaders...)
	}
	headersSet := removeDuplicates(headersList)
	return headersSet, nil
}

func removeDuplicates(values []string) []string {
	keys := make(map[string]struct{})
	var uniqueValues []string
	for _, value := range values {
		if _, ok := keys[value]; !ok {
			keys[value] = struct{}{}
			uniqueValues = append(uniqueValues, value)
		}
	}
	sort.Strings(uniqueValues)
	return uniqueValues
}

func generateFullHeader(prefix, conjunction, suffix string) string {
	if prefix != "" {
		return fmt.Sprintf("%s%s%s", prefix, conjunction, suffix)
	}
	return suffix
}

func escapeCSVInjectionSigns(input string) string {
	if strings.HasPrefix(input, "@") ||
		strings.HasPrefix(input, "=") ||
		strings.HasPrefix(input, "+") ||
		strings.HasPrefix(input, "-") ||
		strings.HasPrefix(input, "\x09") ||
		strings.HasPrefix(input, "\x0D") {
		return fmt.Sprintf("'%s", input)
	}
	return input
}
