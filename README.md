[![Go Reference](https://pkg.go.dev/badge/github.com/simplylib/errgroup.svg)](https://pkg.go.dev/github.com/simplylib/errgroup)
[![Go Report Card](https://goreportcard.com/badge/github.com/simplylib/errgroup)](https://goreportcard.com/report/github.com/simplylib/errgroup)

# [simplylib/errgroup](https://pkg.go.dev/github.com/simplylib/errgroup)
A Go (golang) library that is a drop in replacement for [x/sync/multierror](https://pkg.go.dev/golang.org/x/sync/errgroup) with one major change,

where [x/sync/multierror](https://pkg.go.dev/golang.org/x/sync/errgroup)'s ```Group.Wait()``` "returns the first non-nil error (if any)"

[simplylib/errgroup](https://pkg.go.dev/github.com/simplylib/errgroup)'s ```Group.Wait()``` returns all the errors using [simplylib/multierror](https://pkg.go.dev/github.com/simplylib/multierror)
