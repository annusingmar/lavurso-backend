package helpers

import "golang.org/x/exp/slices"

func VerifyExistsInSlice(toCheck, checkFrom []int) []int {
	var notFound []int

	for _, val := range toCheck {
		e := slices.Contains(checkFrom, val)
		if !e {
			notFound = append(notFound, val)
		}
	}

	return notFound
}
