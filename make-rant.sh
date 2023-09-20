#!/bin/sh

# This script builds DEV0 into executable file in $BUILDDIR.
# Commands are taken from the output of the `go build -n .`.
PROJECT=rant

STDDIR=$HOME/go/pkg

VERBOSITY=0
while test "$1" = "-v"; do
	VERBOSITY=$((VERBOSITY+1))
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

printvv()
{
	if test $VERBOSITY -gt 1; then echo "$@"; fi
}

GOROOT=`go env GOROOT`; export GOROOT

GOMAXPROCS=4; export GOMAXPROCS
GOCACHE=off; export GOCACHE

cat >importcfg << EOF
packagefile runtime=$STDDIR/runtime.a
EOF

cat >importcfg.link << EOF
packagefile internal/abi=$STDDIR/internal/abi.a
packagefile internal/bytealg=$STDDIR/internal/bytealg.a
packagefile internal/coverage/rtcov=$STDDIR/internal/coverage/rtcov.a
packagefile internal/cpu=$STDDIR/internal/cpu.a
packagefile internal/goarch=$STDDIR/internal/goarch.a
packagefile internal/godebugs=$STDDIR/internal/godebugs.a
packagefile internal/goexperiment=$STDDIR/internal/goexperiment.a
packagefile internal/goos=$STDDIR/internal/goos.a
packagefile runtime/internal/atomic=$STDDIR/runtime/internal/atomic.a
packagefile runtime/internal/math=$STDDIR/runtime/internal/math.a
packagefile runtime/internal/sys=$STDDIR/runtime/internal/sys.a
packagefile runtime=$STDDIR/runtime.a
EOF

GOSRC=`ls *.go | grep -v '_test' | grep -v 'stub'`

STARTTIME=`date +%s`

case $1 in
	disas | disasm)
		printvv go tool compile -o $PROJECT.a -S -c=$GOMAXPROCS -nolocalimports -importcfg importcfg -pack $GOSRC
		go tool compile -o $PROJECT.a -S -c=$GOMAXPROCS -nolocalimports -importcfg importcfg -pack $GOSRC >$PROJECT.s 2>&1
		;;
	esc | escape | escape-analysis)
		printvv go tool compile -o $PROJECT.a -m -m -c=$GOMAXPROCS -nolocalimports -importcfg importcfg -pack $GOSRC
		go tool compile -o $PROJECT.a -m -m -c=$GOMAXPROCS -nolocalimports -importcfg importcfg -pack $GOSRC >$PROJECT.esc 2>&1
		;;
	objdump)
		run go tool compile -o $PROJECT.a -p main -c=$GOMAXPROCS -nolocalimports -importcfg importcfg -pack $GOSRC
		run go tool link -o $PROJECT -importcfg=importcfg.link $PROJECT.a
		printvv go tool objdump -S -s ^main\. $PROJECT
		go tool objdump -S -s ^main\. $PROJECT >$PROJECT.s
		;;
	release)
		run go tool compile -o $PROJECT.a -p main -c=$GOMAXPROCS -nolocalimports -importcfg importcfg -pack $GOSRC
		run go tool link -o $PROJECT -s -w -importcfg=importcfg.link $PROJECT.a
		printv "$PROJECT"
		;;
	*)
		echo "Target $1 currently is not supported"
esac

run rm -f importcfg importcfg.link $PROJECT.a

ENDTIME=`date +%s`

echo Done $PROJECT-$1 in $((ENDTIME-STARTTIME))s
