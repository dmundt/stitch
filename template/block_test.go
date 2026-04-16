package template

import "testing"

func TestOrderedBlocks(t *testing.T) {
	got := OrderedBlocks()
	if len(got) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(got))
	}
	if got[0] != BlockHeader || got[1] != BlockMain || got[2] != BlockFooter {
		t.Fatalf("unexpected block order: %#v", got)
	}
}

func TestSortByBlockOrder(t *testing.T) {
	in := []string{BlockFooter, BlockHeader, BlockMain}
	if err := SortByBlockOrder(in); err != nil {
		t.Fatalf("sort failed: %v", err)
	}
	if in[0] != BlockHeader || in[1] != BlockMain || in[2] != BlockFooter {
		t.Fatalf("unexpected sorted order: %#v", in)
	}
}

func TestValidateBlocks(t *testing.T) {
	if err := ValidateBlocks([]string{BlockHeader, "bogus"}); err == nil {
		t.Fatal("expected validation error for invalid block")
	}
}
