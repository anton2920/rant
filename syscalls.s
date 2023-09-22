/* NOTE(anton2920): register order is as follows: rax, rdi, rsi, rdx, r10, r8, r9. */

#include "go_asm.h"
#include "textflag.h"
#include "syscall.h"

TEXT main·Accept(SB), NOSPLIT, $-24
	MOVL s+0(FP), DI
	MOVQ addr+8(FP), SI
	MOVQ addrlen+16(FP), DX
	MOVQ $SYS_accept, AX
	SYSCALL
	MOVL AX, ret+24(FP)
	RET

TEXT main·Bind(SB), NOSPLIT, $-20
	MOVL s+0(FP), DI
	MOVQ addr+8(FP), SI
	MOVL addrlen+16(FP), DX
	MOVQ $SYS_bind, AX
	SYSCALL
	MOVL AX, ret+24(FP)
	RET

TEXT main·Close(SB), NOSPLIT, $-4
	MOVL fd+0(FP), DI
	MOVQ $SYS_close, AX
	SYSCALL
	MOVL AX, ret+4(FP)
	RET

TEXT main·Exit(SB), NOSPLIT, $-4
	MOVL status+0(FP), DI
	MOVQ $SYS_exit, AX
	SYSCALL
	/* NOTE(anton2920): this is the point of noreturn. */

TEXT main·Listen(SB), NOSPLIT, $-8
	MOVL s+0(FP), DI
	MOVL backlog+4(FP), SI
	MOVQ $SYS_listen, AX
	SYSCALL
	MOVL AX, ret+8(FP)
	RET

TEXT main·Socket(SB), NOSPLIT, $-16
	MOVL domain+0(FP), DI
	MOVL type+4(FP), SI
	MOVL proto+12(FP), DX
	MOVQ $SYS_socket, AX
	SYSCALL
	MOVL AX, ret+16(FP)
	RET

TEXT main·Write(SB), NOSPLIT, $-24
	MOVL fd+0(FP), DI
	MOVQ buf+8(FP), SI
	MOVQ n+16(FP), DX
	MOVQ $SYS_write, AX
	SYSCALL
	MOVQ AX, ret+24(FP)
	RET
