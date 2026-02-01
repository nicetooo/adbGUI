package main

import (
	"encoding/json"
	"encoding/xml"
	"testing"
)

// Sample uiautomator XML fragment for testing
const testUINodeXML = `<hierarchy rotation="0">
  <node index="0" text="Hello" resource-id="com.app:id/title" class="android.widget.TextView"
        package="com.app" content-desc="Title label" checkable="false" checked="false"
        clickable="true" enabled="true" focusable="true" focused="false"
        scrollable="false" long-clickable="false" password="false" selected="false"
        bounds="[0,0][540,100]">
    <node index="0" text="" resource-id="com.app:id/icon" class="android.widget.ImageView"
          package="com.app" content-desc="" checkable="true" checked="true"
          clickable="false" enabled="true" focusable="false" focused="false"
          scrollable="true" long-clickable="true" password="true" selected="true"
          bounds="[10,10][50,50]" />
  </node>
</hierarchy>`

func TestUINodeXMLParsing(t *testing.T) {
	var hierarchy UIHierarchy
	err := xml.Unmarshal([]byte(testUINodeXML), &hierarchy)
	if err != nil {
		t.Fatalf("Failed to parse UINode XML: %v", err)
	}

	if len(hierarchy.Nodes) != 1 {
		t.Fatalf("Expected 1 root node, got %d", len(hierarchy.Nodes))
	}

	root := hierarchy.Nodes[0]

	// Test string fields
	if root.Text != "Hello" {
		t.Errorf("Text: expected 'Hello', got %q", root.Text)
	}
	if root.ResourceID != "com.app:id/title" {
		t.Errorf("ResourceID: expected 'com.app:id/title', got %q", root.ResourceID)
	}
	if root.Class != "android.widget.TextView" {
		t.Errorf("Class: expected 'android.widget.TextView', got %q", root.Class)
	}
	if root.Package != "com.app" {
		t.Errorf("Package: expected 'com.app', got %q", root.Package)
	}
	if root.ContentDesc != "Title label" {
		t.Errorf("ContentDesc: expected 'Title label', got %q", root.ContentDesc)
	}
	if root.Bounds != "[0,0][540,100]" {
		t.Errorf("Bounds: expected '[0,0][540,100]', got %q", root.Bounds)
	}

	// Test boolean fields on root node (mixed true/false)
	if root.Checkable != false {
		t.Errorf("Checkable: expected false, got %v", root.Checkable)
	}
	if root.Checked != false {
		t.Errorf("Checked: expected false, got %v", root.Checked)
	}
	if root.Clickable != true {
		t.Errorf("Clickable: expected true, got %v", root.Clickable)
	}
	if root.Enabled != true {
		t.Errorf("Enabled: expected true, got %v", root.Enabled)
	}
	if root.Focusable != true {
		t.Errorf("Focusable: expected true, got %v", root.Focusable)
	}
	if root.Focused != false {
		t.Errorf("Focused: expected false, got %v", root.Focused)
	}
	if root.Scrollable != false {
		t.Errorf("Scrollable: expected false, got %v", root.Scrollable)
	}
	if root.LongClickable != false {
		t.Errorf("LongClickable: expected false, got %v", root.LongClickable)
	}
	if root.Password != false {
		t.Errorf("Password: expected false, got %v", root.Password)
	}
	if root.Selected != false {
		t.Errorf("Selected: expected false, got %v", root.Selected)
	}

	// Test child node with all-true boolean fields
	if len(root.Nodes) != 1 {
		t.Fatalf("Expected 1 child node, got %d", len(root.Nodes))
	}
	child := root.Nodes[0]

	if child.Checkable != true {
		t.Errorf("Child Checkable: expected true, got %v", child.Checkable)
	}
	if child.Checked != true {
		t.Errorf("Child Checked: expected true, got %v", child.Checked)
	}
	if child.Clickable != false {
		t.Errorf("Child Clickable: expected false, got %v", child.Clickable)
	}
	if child.Enabled != true {
		t.Errorf("Child Enabled: expected true, got %v", child.Enabled)
	}
	if child.Focusable != false {
		t.Errorf("Child Focusable: expected false, got %v", child.Focusable)
	}
	if child.Focused != false {
		t.Errorf("Child Focused: expected false, got %v", child.Focused)
	}
	if child.Scrollable != true {
		t.Errorf("Child Scrollable: expected true, got %v", child.Scrollable)
	}
	if child.LongClickable != true {
		t.Errorf("Child LongClickable: expected true, got %v", child.LongClickable)
	}
	if child.Password != true {
		t.Errorf("Child Password: expected true, got %v", child.Password)
	}
	if child.Selected != true {
		t.Errorf("Child Selected: expected true, got %v", child.Selected)
	}
}

func TestUINodeXMLDefaultValues(t *testing.T) {
	// Test node with missing boolean attributes — should default to false
	xmlData := `<hierarchy rotation="0">
  <node index="0" text="Minimal" resource-id="" class="android.view.View"
        package="" content-desc="" bounds="[0,0][100,100]" />
</hierarchy>`

	var hierarchy UIHierarchy
	err := xml.Unmarshal([]byte(xmlData), &hierarchy)
	if err != nil {
		t.Fatalf("Failed to parse minimal UINode XML: %v", err)
	}

	node := hierarchy.Nodes[0]

	// All boolean fields should default to false when not present
	boolFields := map[string]bool{
		"Checkable":     node.Checkable,
		"Checked":       node.Checked,
		"Clickable":     node.Clickable,
		"Enabled":       node.Enabled,
		"Focusable":     node.Focusable,
		"Focused":       node.Focused,
		"Scrollable":    node.Scrollable,
		"LongClickable": node.LongClickable,
		"Password":      node.Password,
		"Selected":      node.Selected,
	}

	for name, val := range boolFields {
		if val != false {
			t.Errorf("%s: expected false (default), got %v", name, val)
		}
	}
}

func TestUINodeJSONSerialization(t *testing.T) {
	node := UINode{
		Text:          "Button",
		ResourceID:    "com.app:id/btn",
		Class:         "android.widget.Button",
		Package:       "com.app",
		Clickable:     true,
		Enabled:       true,
		Focusable:     true,
		LongClickable: false,
		Bounds:        "[0,0][200,80]",
	}

	// Serialize to JSON
	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal UINode: %v", err)
	}

	jsonStr := string(data)

	// Boolean fields should serialize as true/false (not "true"/"false")
	if !contains(jsonStr, `"clickable":true`) {
		t.Errorf("Expected clickable:true in JSON, got %s", jsonStr)
	}
	if !contains(jsonStr, `"enabled":true`) {
		t.Errorf("Expected enabled:true in JSON, got %s", jsonStr)
	}
	if !contains(jsonStr, `"longClickable":false`) {
		t.Errorf("Expected longClickable:false in JSON, got %s", jsonStr)
	}
	if !contains(jsonStr, `"password":false`) {
		t.Errorf("Expected password:false in JSON, got %s", jsonStr)
	}

	// Should NOT have string "true" or "false" (in quotes)
	if contains(jsonStr, `"clickable":"true"`) {
		t.Errorf("Boolean should not be serialized as string: %s", jsonStr)
	}
}

func TestGetNodeAttribute(t *testing.T) {
	app := &App{}
	node := &UINode{
		Text:          "OK",
		ResourceID:    "com.app:id/ok_btn",
		Class:         "android.widget.Button",
		Package:       "com.app",
		ContentDesc:   "Confirm button",
		Bounds:        "[100,200][300,280]",
		Checkable:     false,
		Checked:       false,
		Clickable:     true,
		Enabled:       true,
		Focusable:     true,
		Focused:       false,
		Scrollable:    false,
		LongClickable: true,
		Password:      false,
		Selected:      false,
	}

	// Test string attributes
	tests := []struct {
		attr     string
		expected string
	}{
		{"text", "OK"},
		{"resource-id", "com.app:id/ok_btn"},
		{"resourceid", "com.app:id/ok_btn"},
		{"id", "com.app:id/ok_btn"},
		{"class", "android.widget.Button"},
		{"package", "com.app"},
		{"content-desc", "Confirm button"},
		{"contentdesc", "Confirm button"},
		{"description", "Confirm button"},
		{"desc", "Confirm button"},
		{"bounds", "[100,200][300,280]"},
		// Boolean attributes — should return "true" or "false" strings
		{"clickable", "true"},
		{"enabled", "true"},
		{"focusable", "true"},
		{"focused", "false"},
		{"scrollable", "false"},
		{"checkable", "false"},
		{"checked", "false"},
		{"long-clickable", "true"},
		{"longclickable", "true"},
		{"password", "false"},
		{"selected", "false"},
		// Case insensitivity
		{"Clickable", "true"},
		{"ENABLED", "true"},
		{"Long-Clickable", "true"},
		// Unknown attribute
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		result := app.getNodeAttribute(node, tt.attr)
		if result != tt.expected {
			t.Errorf("getNodeAttribute(%q): expected %q, got %q", tt.attr, tt.expected, result)
		}
	}
}

func TestUINodeNestedParsing(t *testing.T) {
	xmlData := `<hierarchy rotation="0">
  <node index="0" text="Root" resource-id="" class="android.widget.FrameLayout"
        package="com.app" content-desc="" checkable="false" checked="false"
        clickable="false" enabled="true" focusable="false" focused="false"
        scrollable="false" long-clickable="false" password="false" selected="false"
        bounds="[0,0][1080,2400]">
    <node index="0" text="Child1" resource-id="com.app:id/c1" class="android.widget.LinearLayout"
          package="com.app" content-desc="" checkable="false" checked="false"
          clickable="true" enabled="true" focusable="true" focused="true"
          scrollable="false" long-clickable="false" password="false" selected="false"
          bounds="[0,0][1080,1200]">
      <node index="0" text="GrandChild" resource-id="com.app:id/gc" class="android.widget.TextView"
            package="com.app" content-desc="Info" checkable="false" checked="false"
            clickable="false" enabled="true" focusable="false" focused="false"
            scrollable="false" long-clickable="false" password="false" selected="true"
            bounds="[10,10][500,50]" />
    </node>
    <node index="1" text="Child2" resource-id="" class="android.widget.ScrollView"
          package="com.app" content-desc="" checkable="false" checked="false"
          clickable="false" enabled="true" focusable="false" focused="false"
          scrollable="true" long-clickable="false" password="false" selected="false"
          bounds="[0,1200][1080,2400]" />
  </node>
</hierarchy>`

	var hierarchy UIHierarchy
	err := xml.Unmarshal([]byte(xmlData), &hierarchy)
	if err != nil {
		t.Fatalf("Failed to parse nested UINode XML: %v", err)
	}

	root := hierarchy.Nodes[0]
	if root.Text != "Root" {
		t.Errorf("Root text: expected 'Root', got %q", root.Text)
	}
	if len(root.Nodes) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(root.Nodes))
	}

	child1 := root.Nodes[0]
	if child1.Clickable != true {
		t.Error("Child1 should be clickable")
	}
	if child1.Focused != true {
		t.Error("Child1 should be focused")
	}
	if len(child1.Nodes) != 1 {
		t.Fatalf("Child1 expected 1 grandchild, got %d", len(child1.Nodes))
	}

	grandchild := child1.Nodes[0]
	if grandchild.Selected != true {
		t.Error("GrandChild should be selected")
	}
	if grandchild.ContentDesc != "Info" {
		t.Errorf("GrandChild contentDesc: expected 'Info', got %q", grandchild.ContentDesc)
	}

	child2 := root.Nodes[1]
	if child2.Scrollable != true {
		t.Error("Child2 should be scrollable")
	}
}

// contains is a helper for string search
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
