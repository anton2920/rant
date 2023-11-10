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
		CGO_ENABLED=1; export CGO_ENABLED
		run go build -o $PROJECT -race -gcflags='all=-N -l -d=checkptr=0'
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
	prof | profile)
		cp main.go main_back
		sed -e '3a\'$'\n''import _ "net/http/pprof"' -e '3a\'$'\n''import "net/http"' -e '/func main/a\'$'\n''go http.ListenAndServe(":9090", nil)' main.go >new_main
		mv new_main main.go

		run go build -o $PROJECT -asmflags="-I /usr/include" -ldflags='-s -w'
		mv main_back main.go

		echo "Profiling for 60 seconds..."
		./$PROJECT &
		PID=$!
		run curl -o cpu.pprof "http://localhost:9090/debug/pprof/profile?seconds=60" 2>/dev/null
		kill $PID
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
