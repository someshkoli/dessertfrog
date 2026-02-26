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
	buildNodesRecursive("", "", data, 0, &nodes, true, nil)

	return nodes, true
}

// buildNodePath creates a unique path identifier for a node
func buildNodePath(parentPath, key string) string {
	if parentPath == "" {
		return key
	}
	if key != "" {
		return parentPath + "." + key
	}
	return parentPath
}

// buildNodesRecursive recursively builds the JSON tree nodes
// If expansionMap is nil, uses the expanded parameter directly
// If expansionMap is provided, looks up expansion state from the map
func buildNodesRecursive(key, path string, value interface{}, depth int, nodes *[]JSONNode, expanded bool, expansionMap map[string]bool) {
	nodePath := buildNodePath(path, key)

	// Determine expansion state
	nodeExpanded := expanded
	if expansionMap != nil {
		nodeExpanded = expanded && expansionMap[nodePath]
	}

	switch v := value.(type) {
	case map[string]interface{}:
		// Object
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        "object",
			Expanded:    nodeExpanded,
			Depth:       depth,
			HasChildren: len(v) > 0,
		})

		if nodeExpanded && len(v) > 0 {
			for k, val := range v {
				// Pass false for children so they're collapsed by default
				childExpanded := false
				if expansionMap != nil {
					childExpanded = true // When rebuilding, respect expansion map
				}
				buildNodesRecursive(k, nodePath, val, depth+1, nodes, childExpanded, expansionMap)
			}
		}

	case []interface{}:
		// Array
		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        "array",
			Expanded:    nodeExpanded,
			Depth:       depth,
			HasChildren: len(v) > 0,
		})

		if nodeExpanded && len(v) > 0 {
			for i, val := range v {
				// Pass false for children so they're collapsed by default
				childExpanded := false
				if expansionMap != nil {
					childExpanded = true // When rebuilding, respect expansion map
				}
				buildNodesRecursive(fmt.Sprintf("[%d]", i), nodePath, val, depth+1, nodes, childExpanded, expansionMap)
			}
		}

	default:
		// Primitives: string, float64, bool, nil
		nodeType := "string"
		switch v.(type) {
		case float64:
			nodeType = "number"
		case bool:
			nodeType = "boolean"
		case nil:
			nodeType = "null"
		}

		*nodes = append(*nodes, JSONNode{
			Key:         key,
			Value:       v,
			Type:        nodeType,
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
	expansionMap := buildExpansionMap(oldTree)

	// Get the root value and rebuild
	rootNode := oldTree[0]
	var newTree []JSONNode

	buildNodesRecursive("", "", rootNode.Value, 0, &newTree, true, expansionMap)

	return newTree
}

// buildExpansionMap builds a map of node paths to their expansion states
func buildExpansionMap(tree []JSONNode) map[string]bool {
	expansionMap := make(map[string]bool)
	var currentPath string

	for i := 0; i < len(tree); i++ {
		node := tree[i]

		// Calculate path based on depth changes
		if i > 0 {
			prevDepth := tree[i-1].Depth
			currDepth := node.Depth

			if currDepth <= prevDepth {
				// Moving up or same level - need to reconstruct parent path
				currentPath = reconstructPath(tree, i)
			} else {
				// Moving down - append to current path
				currentPath = buildNodePath(currentPath, node.Key)
			}
		} else {
			currentPath = node.Key
		}

		if node.HasChildren {
			expansionMap[currentPath] = node.Expanded
		}
	}

	return expansionMap
}

// reconstructPath reconstructs the full path to a node by walking back through parents
func reconstructPath(tree []JSONNode, index int) string {
	if index == 0 {
		return tree[0].Key
	}

	node := tree[index]
	targetDepth := node.Depth

	// Build path by finding all ancestors
	var pathParts []string
	pathParts = append(pathParts, node.Key)

	// Walk backwards to find parent nodes
	for i := index - 1; i >= 0; i-- {
		if tree[i].Depth == targetDepth-1 {
			pathParts = append([]string{tree[i].Key}, pathParts...)
			targetDepth--
			if targetDepth == 0 {
				break
			}
		}
	}

	// Join path parts
	result := ""
	for i, part := range pathParts {
		if i == 0 {
			result = part
		} else {
			result = buildNodePath(result, part)
		}
	}
	return result
}
