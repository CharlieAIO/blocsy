package dex

import (
	"blocsy/internal/types"
)

func FindTransfer(transfers []types.SolTransfer, innerIndex int, ixIndex int) (*types.SolTransfer, bool) {
	for i := range transfers {
		if transfers[i].IxIndex == ixIndex && transfers[i].InnerIndex == innerIndex {
			return &transfers[i], true
		}
	}
	return nil, false
}

func removeTransfer(transfers []types.SolTransfer, innerIndex int) []types.SolTransfer {
	//for i := len(transfers) - 1; i >= 0; i-- {
	//	if transfers[i].InnerIndex == innerIndex {
	//		transfers = append(transfers[:i], transfers[i+1:]...)
	//	}
	//}
	return transfers
}
