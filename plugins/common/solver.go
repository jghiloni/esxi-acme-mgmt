package common

import "github.com/mholt/acmez/v3"

type Provider interface {
	acmez.Solver
	Name() string
	WithArgs([]string)
}
