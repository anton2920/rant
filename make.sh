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
		run go build -o rant -race -gcflags='all=-N -l' -asmflags='-I /usr/include'
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
	fmt)
		if which goimports >/dev/null; then
			run goimports -l -w *.go
		else
			run gofmt -l -s -w *.go
		fi
		;;
	vet)
		run go vet
		run $0 $VERBOSITYFLAGS clean
		echo "Don't forget to clean up `go env GOCACHE` directory!"
		;;
	disas | disasm | esc | escape | escape-analysis | objdump | release)
		run ./make-rant.sh $VERBOSITYFLAGS $1
		;;
esac

ENDTIME=`date +%s`

echo Done $1 in $((ENDTIME-STARTTIME))s
