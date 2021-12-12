package files

import (
	"fmt"
	"testing"
)

func TestLen(t *testing.T) {
	testSlice := keyIdxSlice([]uint32{1, 2, 3, 4, 5})

	if len := testSlice.Len(); len != 5 {
		t.Errorf("unexpected length: %d", len)
	}
}

func TestLess(t *testing.T) {
	testSlice := keyIdxSlice([]uint32{1, 2, 3, 4, 5})
	tt := []struct {
		left     int
		right    int
		expected bool
	}{
		{1, 4, true},
		{4, 1, false},
		{1, 1, false},
	}

	for _, tc := range tt {
		tc := tc
		name := fmt.Sprintf(
			"%d %s %d",
			tc.left,
			map[bool]string{true: "less than", false: "not less than"}[tc.expected],
			tc.right,
		)
		t.Run(name, func(t *testing.T) {
			if testSlice.Less(tc.left, tc.right) != tc.expected {
				t.Error("unexpected result")
			}
		})
	}
}

func TestSwap(t *testing.T) {
	testSlice := keyIdxSlice([]uint32{1, 2, 3, 4, 5})
	tt := []struct {
		left     int
		right    int
		expected uint32
	}{
		{1, 4, 2},
		{2, 3, 3},
		{3, 4, 3},
	}

	for _, tc := range tt {
		tc := tc
		name := fmt.Sprintf(
			"swapping %d and %d",
			tc.left,
			tc.right,
		)
		t.Run(name, func(t *testing.T) {
			testSlice.Swap(tc.left, tc.right)
			if testSlice[tc.right] != tc.expected {
				t.Errorf(
					"unexpected result\nExpected: %d\nReceived: %d",
					tc.expected,
					[]uint32(testSlice)[tc.right],
				)
			}
		})
	}
}

type fakeKey struct{ epoch uint32 }

func (k *fakeKey) Type() uint32    { return 99 }
func (k *fakeKey) Version() uint32 { return k.epoch }
func (k *fakeKey) Generate() error { return nil }
func (k *fakeKey) Secret() []byte  { return []byte("foo") }
func (k *fakeKey) String() string  { return fmt.Sprintf("fakeKey #%d", k.epoch) }

func TestEachKey(t *testing.T) {
	tt := []struct {
		name    string
		errorAt int
		len     int
		items   []uint32
		err     string
	}{
		{"normal ordering", -1, 5, []uint32{1, 2, 3, 4, 5}, ""},
		{"error at 3", 3, 0, nil, "throwing error at 3"},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			keyring := map[uint32]KeyHandler{
				1: &fakeKey{epoch: 1},
				5: &fakeKey{epoch: 5},
				3: &fakeKey{epoch: 3},
				4: &fakeKey{epoch: 4},
				2: &fakeKey{epoch: 2},
			}
			visited := make([]uint32, 0, len(keyring))
			err := EachKey(keyring, func(id uint32, item KeyHandler) error {
				if tc.errorAt == int(id) {
					return fmt.Errorf("throwing error at %d", tc.errorAt)
				}
				visited = append(visited, id)

				return nil
			})
			if len(tc.err) != 0 {
				if err == nil {
					t.Errorf("unexpected success. Expected: %s", tc.err)
				} else if err.Error() != tc.err {
					t.Errorf("unexpected error.\nExpected: %q\nReceived: %q", tc.err, err.Error())
				}
			}
			if err != nil {
				return
			}
			if len(visited) != tc.len {
				t.Errorf("not all items were visited: %d", tc.len)
			}
			for idx := range tc.items {
				if tc.items[idx] != visited[idx] {
					t.Errorf("item out of order: %d (nth: %d)", visited[idx], idx)
				}
			}
		})
	}
}
