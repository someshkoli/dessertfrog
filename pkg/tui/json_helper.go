package tui

import (
	"encoding/json"
	"fmt"
)

// buildJSONTree parses a JSON string and builds a flat tree structure for display
func buildJSONTree(jsonStr string) ([]JSONNode, bool) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, false
	}

	var nodes []JSONNode
	buildNodesRecursive("", data, 0, &nodes, true)

	return nodes, true
}

// buildNodesRecursive recursively builds the JSON tree nodes
func buildNodesRecursive(key string, value interface{}, depth int, nodes *[]JSONNode, expanded bool) {
	switch v := value.(type) {
	case map[string]interface{}:
		// Object
		node := JSONNode{
			Key:         key,
			Value:       v,
			Type:        "object",
			Expanded:    expanded,
			Depth:       depth,
			HasChildren: len(v) > 0,
		}
		*nodes = append(*nodes, node)

		if expanded && len(v) > 0 {
			for k, val := range v {
				buildNodesRecursive(k, val, depth+1, nodes, false)
			}
		}

	case []interface{}:
		// Array
		node := JSONNode{
			Key:         key,
			Value:       v,
			Type:        "array",
			Expanded:    expanded,
			Depth:       depth,
			HasChildren: len(v) > 0,
		}
		*nodes = append(*nodes, node)

		if expanded && len(v) > 0 {
			for i, val := range v {
				buildNodesRecursive(fmt.Sprintf("[%d]", i), val, depth+1, nodes, false)
			}
		}

	case string:
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        "string",
			Expanded:    false,
			Depth:       depth,
			HasChildren: false,
		})

	case float64:
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        "number",
			Expanded:    false,
			Depth:       depth,
			HasChildren: false,
		})

	case bool:
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        "boolean",
			Expanded:    false,
			Depth:       depth,
			HasChildren: false,
		})

	case nil:
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       nil,
			Type:        "null",
			Expanded:    false,
			Depth:       depth,
			HasChildren: false,
		})
	}
}

// rebuildVisibleJSONTree rebuilds the tree with current expand/collapse states
func rebuildVisibleJSONTree(oldTree []JSONNode) []JSONNode {
	if len(oldTree) == 0 {
		return oldTree
	}

	// Extract expansion states
	expansionMap := make(map[string]bool)
	for _, node := range oldTree {
		if node.HasChildren {
			key := fmt.Sprintf("%d_%s", node.Depth, node.Key)
			expansionMap[key] = node.Expanded
		}
	}

	// Get the root value and rebuild
	rootNode := oldTree[0]
	var newTree []JSONNode

	buildNodesRecursiveWithState("", rootNode.Value, 0, &newTree, true, expansionMap)

	return newTree
}

// buildNodesRecursiveWithState builds nodes while preserving expansion states
func buildNodesRecursiveWithState(key string, value interface{}, depth int, nodes *[]JSONNode, parentExpanded bool, expansionMap map[string]bool) {
	stateKey := fmt.Sprintf("%d_%s", depth, key)
	expanded := parentExpanded && expansionMap[stateKey]

	switch v := value.(type) {
	case map[string]interface{}:
		// Object
		node := JSONNode{
			Key:         key,
			Value:       v,
			Type:        "object",
			Expanded:    expanded,
			Depth:       depth,
			HasChildren: len(v) > 0,
		}
		*nodes = append(*nodes, node)

		if parentExpanded && expanded && len(v) > 0 {
			for k, val := range v {
				buildNodesRecursiveWithState(k, val, depth+1, nodes, true, expansionMap)
			}
		}

	case []interface{}:
		// Array
		node := JSONNode{
			Key:         key,
			Value:       v,
			Type:        "array",
			Expanded:    expanded,
			Depth:       depth,
			HasChildren: len(v) > 0,
		}
		*nodes = append(*nodes, node)

		if parentExpanded && expanded && len(v) > 0 {
			for i, val := range v {
				buildNodesRecursiveWithState(fmt.Sprintf("[%d]", i), val, depth+1, nodes, true, expansionMap)
			}
		}

	case string:
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        "string",
			Expanded:    false,
			Depth:       depth,
			HasChildren: false,
		})

	case float64:
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        "number",
			Expanded:    false,
			Depth:       depth,
			HasChildren: false,
		})

	case bool:
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        "boolean",
			Expanded:    false,
			Depth:       depth,
			HasChildren: false,
		})

	case nil:
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       nil,
			Type:        "null",
			Expanded:    false,
			Depth:       depth,
			HasChildren: false,
		})
	}
}
