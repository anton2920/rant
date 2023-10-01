#!/bin/sh

# This script builds DEV0 into executable file in $BUILDDIR.
# Commands are taken from the output of the `go build -n .`.
PROJECT=rant

BUILDDIR=`pwd`/build
STDDIR=`pwd`/build

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

GOOS=`go env GOOS`; export GOOS
GOARCH=`go env GOARCH`; export GOARCH
GOAMD64=`go env GOAMD64`; export GOAMD64

GOMAXPROCS=4; export GOMAXPROCS
GOCACHE=off; export GOCACHE

cat >$BUILDDIR/importcfg << EOF
packagefile runtime=$STDDIR/runtime.a
EOF

cat >$BUILDDIR/importcfg.link << EOF
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

ASMSRC=`ls *.s | grep -v 'rant'`
GOSRC=`ls *.go | grep -v '_test'`

STARTTIME=`date +%s`

touch $BUILDDIR/go_asm.h

case $1 in
	disas | disasm)
		run go tool asm -p main -I $BUILDDIR -I $GOROOT/pkg/include -I /usr/include -D GOOS_$GOOS -D GOARCH_$GOARCH $5 -D GOAMD64_$GOAMD64  -I $GOROOT/src/runtime -gensymabis -o $BUILDDIR/symabis $ASMSRC
		printvv go tool compile -o $BUILDDIR/main.a -trimpath "$BUILDDIR=>" -p main -symabis $BUILDDIR/symabis -S -c=$GOMAXPROCS -nolocalimports -importcfg $BUILDDIR/importcfg -pack -asmhdr $BUILDDIR/go_asm.h $GOSRC
		go tool compile -o $BUILDDIR/main.a -trimpath "$BUILDDIR=>" -p main -symabis $BUILDDIR/symabis -S -c=$GOMAXPROCS -nolocalimports -importcfg $BUILDDIR/importcfg -pack -asmhdr $BUILDDIR/go_asm.h $GOSRC >$PROJECT.s 2>&1
		;;
	esc | escape | escape-analysis)
		printvv go tool compile -o $BUILDDIR/main.a -trimpath "$BUILDDIR=>" -m -m -c=$GOMAXPROCS -nolocalimports -importcfg $BUILDDIR/importcfg -pack $GOSRC
		go tool compile -o $BUILDDIR/main.a -trimpath "$BUILDDIR=>" -m -m -c=$GOMAXPROCS -nolocalimports -importcfg $BUILDDIR/importcfg -pack $GOSRC >$PROJECT.esc 2>&1
		;;
	objdump)
		run go tool asm -p main -I $BUILDDIR -I $GOROOT/pkg/include -I /usr/include -D GOOS_$GOOS -D GOARCH_$GOARCH $5 -D GOAMD64_$GOAMD64  -I $GOROOT/src/runtime -gensymabis -o $BUILDDIR/symabis $ASMSRC
		run go tool compile -o $BUILDDIR/main.a -trimpath "$BUILDDIR=>" -p main -symabis $BUILDDIR/symabis -c=$GOMAXPROCS -nolocalimports -importcfg $BUILDDIR/importcfg -pack -asmhdr $BUILDDIR/go_asm.h $GOSRC
		for file in $ASMSRC; do
			run go tool asm -p main -I $BUILDDIR -I $GOROOT/pkg/include -I /usr/include -D GOOS_$GOOS -D GOARCH_$GOARCH -D GOAMD64_$GOAMD64 -I $GOROOT/src/runtime -o $BUILDDIR/`basename $file`.o $file
		done
		run go tool pack r $BUILDDIR/main.a $BUILDDIR/*.o
		run go tool link -o $PROJECT -importcfg=$BUILDDIR/importcfg.link $BUILDDIR/main.a
		printvv go tool objdump -S -s ^main\. $PROJECT
		go tool objdump -S -s ^main\. $PROJECT >$PROJECT.s
		;;
	release)
		run go tool asm -p main -I $BUILDDIR -I $GOROOT/pkg/include -I /usr/include -D GOOS_$GOOS -D GOARCH_$GOARCH $5 -D GOAMD64_$GOAMD64  -I $GOROOT/src/runtime -gensymabis -o $BUILDDIR/symabis $ASMSRC
		run go tool compile -o $BUILDDIR/main.a -trimpath "$BUILDDIR=>" -p main -pgoprofile=default.pgo -symabis $BUILDDIR/symabis -c=$GOMAXPROCS -nolocalimports -importcfg $BUILDDIR/importcfg -pack -asmhdr $BUILDDIR/go_asm.h $GOSRC
		for file in $ASMSRC; do
			run go tool asm -p main -I $BUILDDIR -I $GOROOT/pkg/include -I /usr/include/ -D GOOS_$GOOS -D GOARCH_$GOARCH -D GOAMD64_$GOAMD64 -I $GOROOT/src/runtime -o $BUILDDIR/`basename $file`.o $file
		done
		run go tool pack r $BUILDDIR/main.a $BUILDDIR/*.o
		run go tool link -o $PROJECT -s -w -importcfg=$BUILDDIR/importcfg.link $BUILDDIR/main.a
		printv "$PROJECT"
		;;
	*)
		echo "Target $1 currently is not supported"
esac

run rm -f $BUILDDIR/importcfg $BUILDDIR/importcfg.link $BUILDDIR/go_asm.h $BUILDDIR/symabis $BUILDDIR/main.a $BUILDDIR/*.o

ENDTIME=`date +%s`

echo Done $PROJECT-$1 in $((ENDTIME-STARTTIME))s
