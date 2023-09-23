/* NOTE(anton2920): register order is as follows: rax, rdi, rsi, rdx, r10, r8, r9. */

#include "go_asm.h"
#include "textflag.h"
#include "syscall.h"

/* func Accept(s int32, addr *SockAddr, addrlen *uint32) int32 */
TEXT main·Accept(SB), NOSPLIT, $0-20
	MOVQ $SYS_accept, AX
	MOVL s+0(FP), DI
	MOVQ addr+8(FP), SI
	MOVQ paddrlen+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+24(FP)
	RET

/* func Bind(s int32, addr *SockAddr, addrlen uint32) int32 */
TEXT main·Bind(SB), NOSPLIT, $0-16
	MOVQ $SYS_bind, AX
	MOVL s+0(FP), DI
	MOVQ addr+8(FP), SI
	MOVL addrlen+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+24(FP)
	RET

/* func Close(fd int32) int32 */
TEXT main·Close(SB), NOSPLIT, $0-4
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
	/* NOTE(anton2920): this is the point of noreturn. */
	MOVQ 0x0, AX

/* func Kevent(kq int32, changelist []Kevent, eventlist []Kevent, timeout *Timespec) */
TEXT main·Kevent(SB), NOSPLIT, $0-60
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

/* func Kqueue() int32 */
TEXT main·Kqueue(SB), NOSPLIT, $0-0
	MOVQ $SYS_kqueue, AX
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+0(FP)
	RET

/* func Listen(s int32, backlog int32) int32 */
TEXT main·Listen(SB), NOSPLIT, $0-8
	MOVQ $SYS_listen, AX
	MOVL s+0(FP), DI
	MOVL backlog+4(FP), SI
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+8(FP)
	RET

/* func Lseek(fd int32, offset int, whence int32) int */
TEXT main·Lseek(SB), NOSPLIT, $0-16
	MOVQ $SYS_lseek, AX
	MOVL fd+0(FP), DI
	MOVQ offset+8(FP), SI
	MOVQ whence+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVQ AX, ret+24(FP)
	RET

/* func Nanosleep(rqtp, rmtp *Timespec) int32 */
TEXT main·Nanosleep(SB), NOSPLIT, $0-16
	MOVQ $SYS_nanosleep, AX
	MOVQ rqtp+0(FP), DI
	MOVQ rmtp+8(FP), SI
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+16(FP)
	RET

/* func Open(path string, flags int32, mode uint16) int32 */
TEXT main·Open(SB), NOSPLIT, $0-22
	MOVQ $SYS_open, AX
	MOVQ path+0(FP), DI
	MOVL flags+16(FP), SI
	MOVW mode+20(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVQ AX, ret+24(FP)
	RET

/* func Read(fd int32, buf []byte) int */
TEXT main·Read(SB), NOSPLIT, $0-28
	MOVQ $SYS_read, AX
	MOVL fd+0(FP), DI
	MOVQ buf+8(FP), SI
	MOVQ n+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVQ AX, ret+32(FP)
	RET

/* func Setsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) int32 */
TEXT main·Setsockopt(SB), NOSPLIT, $0-24
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

/* func Shutdown(s int32, how int32) int32 */
TEXT main·Shutdown(SB), NOSPLIT, $0-8
	MOVQ $SYS_shutdown, AX
	MOVL s+0(FP), DI
	MOVL how+4(FP), SI
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVL AX, ret+8(FP)
	RET

/* func Socket(domain, typ, protocol int32) int32 */
TEXT main·Socket(SB), NOSPLIT, $0-12
	MOVQ $SYS_socket, AX
	MOVL domain+0(FP), DI
	MOVL type+4(FP), SI
	MOVL proto+8(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGL AX
	MOVL AX, ret+16(FP)
	RET

/* func Write(fd int32, buf []byte) int */
TEXT main·Write(SB), NOSPLIT, $0-28
	MOVQ $SYS_write, AX
	MOVL fd+0(FP), DI
	MOVQ buf+8(FP), SI
	MOVQ n+16(FP), DX
	SYSCALL
	JCC 2(PC)
	NEGQ AX
	MOVQ AX, ret+32(FP)
	RET
