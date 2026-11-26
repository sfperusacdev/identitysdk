package setup

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

func TestService_readProperties_Hidden(t *testing.T) {
	tests := []struct {
		name   string
		hidden string
		want   bool
	}{
		{name: "omitted", want: false},
		{name: "visible", hidden: "<hidden>false</hidden>", want: false},
		{name: "hidden", hidden: "<hidden>true</hidden>", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsys := fstest.MapFS{
				"_properties/properties.xml": &fstest.MapFile{
					Data: []byte("<SystemProperties><SystemProperty>" +
						"<id>example</id><title>Example</title><type>string</type>" +
						tt.hidden + "<value>configured</value>" +
						"</SystemProperty></SystemProperties>"),
				},
			}

			entries, err := (&Service{}).readProperties(fsys)
			if err != nil {
				t.Fatalf("readProperties() error = %v", err)
			}
			if len(entries) != 1 || entries[0].Hidden != tt.want {
				t.Fatalf("readProperties() = %+v, want one entry with hidden=%t", entries, tt.want)
			}

			data, err := json.Marshal(entries[0])
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if hidden, ok := result["hidden"]; !ok || hidden != tt.want {
				t.Errorf("JSON hidden = %v (present: %t), want %t", hidden, ok, tt.want)
			}
		})
	}
}
