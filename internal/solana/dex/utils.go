package dex

import "defi-intel/internal/types"

func FindTransfer(transfers []types.SolTransfer, innerIndex int, ixIndex int) (*types.SolTransfer, bool) {
	for i := range transfers {
		if transfers[i].IxIndex == ixIndex && transfers[i].InnerIndex == innerIndex {
			return &transfers[i], true
		}
	}
	return nil, false
}
