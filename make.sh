#!/bin/sh

PROJECT=rant

VERBOSITY=0
VERBOSITYFLAGS=""
while test "$1" = "-v"; do
	VERBOSITY=$((VERBOSITY+1))
	VERBOSITYFLAGS="$VERBOSITYFLAGS -v"
	shift
done

run()
{
	if test $VERBOSITY -gt 1; then echo "$@"; fi
	"$@" || exit 1
}

printv()
{
	if test $VERBOSITY -gt 0; then echo "$@"; fi
}

# NOTE(anton2920): don't like Google spying on me.
GOPROXY=direct; export GOPROXY
GOSUMDB=off; export GOSUMDB

# NOTE(anton2920): disable Go 1.11+ package management.
GO111MODULE=off; export GO111MODULE
GOPATH=`go env GOPATH`:`pwd`/vendor; export GOPATH

STARTTIME=`date +%s`

case $1 in
	'' | debug)
		run go build $VERBOSITYFLAGS -o $PROJECT -race -gcflags='all=-N -l' .
		echo "Don't forget to clean up `go env GOCACHE` directory!"
		;;
	all)
		printv "Building Go standard library..."
		run ./make-std.sh $VERBOSITYFLAGS
		run $0 $VERBOSITYFLAGS release
		;;
	clean)
		run rm -f $PROJECT $PROJECT.s $PROJECT.esc $PROJECT.test c.out cpu.pprof mem.pprof
		run go clean -cache -modcache -testcache
		run rm -rf `go env GOCACHE`
		run rm -rf /tmp/cover*
		;;
	check)
		run go vet
		check_db_variable
		run go test -race -cover
		echo "Don't forget to clean up `go env GOCACHE` directory!"
		;;
	check-bench)
		export CGO_ENABLED=0
		run go vet
		check_db_variable
		run go test -bench=. -run=^Benchmark -benchmem
		echo "Don't forget to clean up `go env GOCACHE` directory!"
		;;
	check-bench-cpu)
		export CGO_ENABLED=0
		run go vet
		check_db_variable
		run go test -v -bench=. -run=^Benchmark -benchmem -cpuprofile=cpu.pprof
		echo "Don't forget to clean up `go env GOCACHE` directory!"
		;;
	check-bench-mem)
		export CGO_ENABLED=0
		run go vet
		check_db_variable
		run go test -v -bench=. -run=^Benchmark -benchmem -cpuprofile=mem.pprof
		echo "Don't forget to clean up `go env GOCACHE` directory!"
		;;
	check-cover | check-cover-report)
		run go vet
		check_db_variable
		run go test -race -coverprofile=c.out
		run go tool cover -html=c.out
		run rm -f c.out
		echo "Don't forget to clean up `go env GOCACHE` directory!"
		;;
	fmt)
		if which goimports >/dev/null; then
			run goimports -l -w *.go
		else
			run gofmt -l -s -w *.go
		fi
		;;
	disas | disasm | esc | escape | escape-analysis | objdump | release)
		run ./make-rant.sh $VERBOSITYFLAGS $1
		;;
esac

ENDTIME=`date +%s`

echo Done $1 in $((ENDTIME-STARTTIME))s
