package macho

import (
	"strings"
	"testing"
)

func TestSectionSizeChangesLargestFirst(t *testing.T) {
	oldInfo := &DiffInfo{Sections: []section{
		{Name: "__TEXT.__text", Size: 0x100},
		{Name: "__DATA.__data", Size: 0x80},
		{Name: "__LINKEDIT", Size: 0x400},
	}}
	newInfo := &DiffInfo{Sections: []section{
		{Name: "__TEXT.__text", Size: 0x180},
		{Name: "__DATA.__data", Size: 0x70},
		{Name: "__NEW.__foo", Size: 0x200},
	}}
	got := sectionSizeChanges(oldInfo, newInfo)
	want := []string{"__LINKEDIT", "__NEW.__foo", "__TEXT.__text", "__DATA.__data"}
	if len(got) != len(want) {
		t.Fatalf("got %d changes, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Name != want[i] {
			t.Fatalf("change[%d] = %s, want %s", i, got[i].Name, want[i])
		}
	}
}

func TestFormatUpdatedDiffMarkdownPrioritizesSizeChanges(t *testing.T) {
	oldInfo := &DiffInfo{
		Sections: []section{
			{Name: "__TEXT.__text", Size: 0x100, Hash: "same", HashMode: hashRaw},
			{Name: "__TEXT.__const", Size: 0x40, Hash: "old", HashMode: hashRaw},
		},
		Functions: 10,
	}
	newInfo := &DiffInfo{
		Sections: []section{
			{Name: "__TEXT.__text", Size: 0x120, Hash: "same", HashMode: hashRaw},
			{Name: "__TEXT.__const", Size: 0x40, Hash: "new", HashMode: hashRaw},
		},
		Functions: 11,
	}
	got, err := FormatUpdatedDiff(oldInfo, newInfo, &DiffConfig{Markdown: true})
	if err != nil {
		t.Fatal(err)
	}
	sizePos := strings.Index(got, "### Section Size Changes")
	contentPos := strings.Index(got, "### Same-size Content Changes")
	otherPos := strings.Index(got, "### Other Changes")
	if sizePos < 0 || contentPos < 0 || otherPos < 0 {
		t.Fatalf("missing expected sections:\n%s", got)
	}
	if !(sizePos < contentPos && contentPos < otherPos) {
		t.Fatalf("unexpected section order:\n%s", got)
	}
	if !strings.Contains(got, "| `__TEXT.__text` | `0x100` | `0x120` | **`+0x20`** |") {
		t.Fatalf("missing size delta table row:\n%s", got)
	}
	if !strings.Contains(got, "- `__TEXT.__const`") {
		t.Fatalf("missing same-size content change:\n%s", got)
	}
}
