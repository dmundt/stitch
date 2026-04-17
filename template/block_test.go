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

func TestValidateBlocksEmpty(t *testing.T) {
	if err := ValidateBlocks([]string{}); err != nil {
		t.Fatalf("empty block list should be valid: %v", err)
	}
}

func TestIsValidBlock(t *testing.T) {
	for _, name := range []string{BlockHeader, BlockMain, BlockFooter} {
		if !IsValidBlock(name) {
			t.Fatalf("expected %q to be a valid block", name)
		}
	}
	if IsValidBlock("sidebar") {
		t.Fatal("expected 'sidebar' to be invalid")
	}
	if IsValidBlock("") {
		t.Fatal("expected empty string to be invalid")
	}
}

func TestSortByBlockOrderInvalidBlock(t *testing.T) {
	err := SortByBlockOrder([]string{BlockHeader, "bogus"})
	if err == nil {
		t.Fatal("expected error for invalid block in sort")
	}
}
