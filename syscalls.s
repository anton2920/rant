#include "go_asm.h"
#include "textflag.h"

/* NOTE(anton2920): definitions are taken from <sys/syscall.h>. */
#define SYS_exit   1
#define SYS_write  4
#define SYS_socket 97

/* NOTE(anton2920): register order is as follows: rax, rdi, rsi, rdx, r10, r8, r9. */

TEXT main·Exit(SB), NOSPLIT, $8-0
	MOVL status+0(FP), DI
	MOVQ $SYS_exit, AX
	SYSCALL
	/* NOTE(anton2920): this is the point of noreturn. */

TEXT main·Socket(SB), NOSPLIT, $24-0
	MOVL domain+0(FP), DI
	MOVL type+0(FP), SI
	MOVL proto+0(FP), DX
	MOVQ $SYS_socket, AX
	SYSCALL
	RET

TEXT main·Write(SB), NOSPLIT, $24-0
	MOVL fd+0(FP), DI
	MOVQ buf+8(FP), SI
	MOVQ n+16(FP), DX
	MOVQ $SYS_write, AX
	SYSCALL
	RET
