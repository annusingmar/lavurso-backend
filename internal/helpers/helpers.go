package helpers

import (
	"github.com/go-jet/jet/v2/postgres"
	"golang.org/x/exp/slices"
)

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

func ToPtr[T any](v T) *T {
	return &v
}

func PostgresInt(i int) postgres.IntegerExpression {
	return postgres.Int32(int32(i))
}
