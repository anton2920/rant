/* NOTE(anton2920): register order is as follows: rax, rdi, rsi, rdx, r10, r8, r9. */

#include "go_asm.h"
#include "textflag.h"
#include "sys/syscall.h"

/* func RawAccept(s int32, addr *SockAddr, addrlen *uint32) int32 */
TEXT main·RawAccept(SB), NOSPLIT, $0-20
	MOVQ $SYS_accept, AX
	MOVL s+0(FP), DI
	MOVQ addr+8(FP), SI
	MOVQ paddrlen+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+24(FP)
	RET

/* func RawBind(s int32, addr *SockAddr, addrlen uint32) int32 */
TEXT main·RawBind(SB), NOSPLIT, $0-16
	MOVQ $SYS_bind, AX
	MOVL s+0(FP), DI
	MOVQ addr+8(FP), SI
	MOVL addrlen+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+24(FP)
	RET

/* func RawClose(fd int32) int32 */
TEXT main·RawClose(SB), NOSPLIT, $0-4
	MOVQ $SYS_close, AX
	MOVL fd+0(FP), DI
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+8(FP)
	RET

/* func Exit(status int32) */
TEXT main·Exit(SB), NOSPLIT, $0-4
	MOVQ $SYS_exit, AX
	MOVL status+0(FP), DI
	SYSCALL
	/* NOTE(anton2920): this is the point64 of noreturn. */
	MOVQ 0x0, AX

/* func RawFstat(fd int32, sb *Stat) int32 */
TEXT main·RawFstat(SB), NOSPLIT, $0-12
	MOVQ $SYS_fstat, AX
	MOVL fd+0(FP), DI
	MOVQ sb+8(FP), SI
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+16(FP)
	RET

/* func RawKevent(kq int32, changelist []Kevent, eventlist []Kevent, timeout *Timespec) int32 */
TEXT main·RawKevent(SB), NOSPLIT, $0-60
	MOVQ $SYS_kevent, AX
	MOVL kq+0(FP), DI
	MOVQ chlist+8(FP), SI
	MOVQ nchanges+16(FP), DX
	MOVQ evlist+32(FP), R10
	MOVQ nevents+40(FP), R8
	MOVQ timeout+56(FP), R9
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+64(FP)
	RET

/* func RawKqueue() int32 */
TEXT main·RawKqueue(SB), NOSPLIT, $0-0
	MOVQ $SYS_kqueue, AX
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+0(FP)
	RET

/* func RawListen(s int32, backlog int32) int32 */
TEXT main·RawListen(SB), NOSPLIT, $0-8
	MOVQ $SYS_listen, AX
	MOVL s+0(FP), DI
	MOVL backlog+4(FP), SI
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+8(FP)
	RET

/* func RawLseek(fd int32, offset int64, whence int32) int64 */
TEXT main·RawLseek(SB), NOSPLIT, $0-16
	MOVQ $SYS_lseek, AX
	MOVL fd+0(FP), DI
	MOVQ offset+8(FP), SI
	MOVQ whence+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVQ AX, ret+24(FP)
	RET

/* func RawNanosleep(rqtp, rmtp *Timespec) int32 */
TEXT main·RawNanosleep(SB), NOSPLIT, $0-16
	MOVQ $SYS_nanosleep, AX
	MOVQ rqtp+0(FP), DI
	MOVQ rmtp+8(FP), SI
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+16(FP)
	RET

/* func RawOpen(path string, flags int32, mode uint16) int32 */
TEXT main·RawOpen(SB), NOSPLIT, $0-22
	MOVQ $SYS_open, AX
	MOVQ path+0(FP), DI
	MOVL flags+16(FP), SI
	MOVW mode+20(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVQ AX, ret+24(FP)
	RET

/* func RawRead(fd int32, buf []byte) int64 */
TEXT main·RawRead(SB), NOSPLIT, $0-28
	MOVQ $SYS_read, AX
	MOVL fd+0(FP), DI
	MOVQ buf+8(FP), SI
	MOVQ n+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVQ AX, ret+32(FP)
	RET

/* func RawSetsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) int32 */
TEXT main·RawSetsockopt(SB), NOSPLIT, $0-24
	MOVQ $SYS_setsockopt, AX
	MOVL s+0(FP), DI
	MOVL lvl+4(FP), SI
	MOVL opt+8(FP), DX
	MOVQ val+16(FP), R10
	MOVL len+24(FP), R8
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+32(FP)
	RET

/* func RawShutdown(s int32, how int32) int32 */
TEXT main·RawShutdown(SB), NOSPLIT, $0-8
	MOVQ $SYS_shutdown, AX
	MOVL s+0(FP), DI
	MOVL how+4(FP), SI
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVL AX, ret+8(FP)
	RET

/* func RawSocket(domain, typ, protocol int32) int32 */
TEXT main·RawSocket(SB), NOSPLIT, $0-12
	MOVQ $SYS_socket, AX
	MOVL domain+0(FP), DI
	MOVL type+4(FP), SI
	MOVL proto+8(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+16(FP)
	RET

/* func RawWrite(fd int32, buf []byte) int64 */
TEXT main·RawWrite(SB), NOSPLIT, $0-28
	MOVQ $SYS_write, AX
	MOVL fd+0(FP), DI
	MOVQ buf+8(FP), SI
	MOVQ n+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVQ AX, ret+32(FP)
	RET
