package vdom

import (
	"testing"
)

func TestCreateElementTextWithStyle(t *testing.T) {
	props := map[string]any{
		"style": map[string]any{
			"color": "red",
			"bold":  true,
		},
	}

	node := CreateElement("text", props, "hello")

	textNode, ok := node.(*TextNode)
	if !ok {
		t.Fatalf("Expected TextNode, got %T", node)
	}

	if textNode.Content != "hello" {
		t.Errorf("Expected content 'hello', got '%s'", textNode.Content)
	}

	if textNode.Style.Foreground != "red" {
		t.Errorf("Expected Style.Foreground 'red', got '%s'", textNode.Style.Foreground)
	}

	if !textNode.Style.Bold {
		t.Errorf("Expected Style.Bold true, got false")
	}
}
