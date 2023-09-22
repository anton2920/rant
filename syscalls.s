/* NOTE(anton2920): register order is as follows: rax, rdi, rsi, rdx, r10, r8, r9. */

#include "go_asm.h"
#include "textflag.h"
#include "syscall.h"

/* func Accept(s int32, addr *SockAddr, addrlen *uint32) int32 */
TEXT main·Accept(SB), NOSPLIT, $0-20
	MOVL s+0(FP), DI
	MOVQ addr+8(FP), SI
	MOVQ paddrlen+16(FP), DX
	MOVQ $SYS_accept, AX
	SYSCALL
	MOVL AX, ret+24(FP)
	RET

/* func Bind(s int32, addr *SockAddr, addrlen uint32) int32 */
TEXT main·Bind(SB), NOSPLIT, $0-16
	MOVL s+0(FP), DI
	MOVQ addr+8(FP), SI
	MOVL addrlen+16(FP), DX
	MOVQ $SYS_bind, AX
	SYSCALL
	MOVL AX, ret+24(FP)
	RET

/* func Close(fd int32) int32 */
TEXT main·Close(SB), NOSPLIT, $0-4
	MOVL fd+0(FP), DI
	MOVQ $SYS_close, AX
	SYSCALL
	MOVL AX, ret+8(FP)
	RET

/* func Exit(status int32) */
TEXT main·Exit(SB), NOSPLIT, $0-4
	MOVL status+0(FP), DI
	MOVQ $SYS_exit, AX
	SYSCALL
	/* NOTE(anton2920): this is the point of noreturn. */

/* func Listen(s int32, backlog int32) int32 */
TEXT main·Listen(SB), NOSPLIT, $0-8
	MOVL s+0(FP), DI
	MOVL backlog+4(FP), SI
	MOVQ $SYS_listen, AX
	SYSCALL
	MOVL AX, ret+8(FP)
	RET

/* func Setsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) int32 */
TEXT main·Setsockopt(SB), NOSPLIT, $0-24
	MOVL s+0(FP), DI
	MOVL lvl+4(FP), SI
	MOVL opt+8(FP), DX
	MOVQ val+16(FP), R10
	MOVL len+24(FP), R8
	MOVQ $SYS_setsockopt, AX
	SYSCALL
	MOVL AX, ret+32(FP)
	RET

/* func Socket(domain, typ, protocol int32) int32 */
TEXT main·Socket(SB), NOSPLIT, $0-12
	MOVL domain+0(FP), DI
	MOVL type+4(FP), SI
	MOVL proto+8(FP), DX
	MOVQ $SYS_socket, AX
	SYSCALL
	MOVL AX, ret+16(FP)
	RET

/* func Write(fd int32, buf []byte) int */
TEXT main·Write(SB), NOSPLIT, $0-28
	MOVL fd+0(FP), DI
	MOVQ buf+8(FP), SI
	MOVQ n+16(FP), DX
	MOVQ $SYS_write, AX
	SYSCALL
	MOVQ AX, ret+32(FP)
	RET
