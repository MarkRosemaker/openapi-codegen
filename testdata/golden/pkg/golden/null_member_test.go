package golden

import (
	"encoding/json/v2"
	"testing"
)

// TestNode_NullMemberIsNotThere decodes a folder, whose file is null, strictly and leniently: both must pick the
// alternative that allows a null file, not the first that requires one.
func TestNode_NullMemberIsNotThere(t *testing.T) {
	raw := []byte(`{"name":"docs","folder":{"count":2},"file":null}`)

	for name, opts := range map[string]json.Options{"strict": jsonOpts, "lenient": jsonOptsLenient} {
		t.Run(name, func(t *testing.T) {
			var n Node
			if err := json.Unmarshal(raw, &n, opts); err != nil {
				t.Fatal(err)
			}

			if n.FolderDetails == nil || n.FileDetails != nil {
				t.Fatalf("got file %v and folder %v, want the folder alone", n.FileDetails, n.FolderDetails)
			}

			if n.FolderDetails.Folder.Count != 2 {
				t.Errorf("count = %d, want 2", n.FolderDetails.Folder.Count)
			}
		})
	}
}
