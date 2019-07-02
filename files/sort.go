package files

import "sort"

type keyIdxSlice []uint32

func (p keyIdxSlice) Len() int           { return len(p) }
func (p keyIdxSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p keyIdxSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// EachKey loops over keys by ascending order by epoch number
func EachKey(keys map[uint32]KeyHandler, callback func(uint32, KeyHandler) error) error {
	index := make([]uint32, 0, len(keys))
	for k := range keys {
		index = append(index, k)
	}
	sort.Sort(keyIdxSlice(index))
	for _, k := range index {
		if err := callback(k, keys[k]); err != nil {
			return err
		}
	}
	return nil
}
