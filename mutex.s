#include "go_asm.h"
#include "textflag.h"
#include "sys/syscall.h"

/* func Cas32(val *int32, old, new int32) bool */
TEXT main·Cas32(SB), NOSPLIT, $0-16
	MOVQ ptr+0(FP), BX
	MOVL old+8(FP), AX
	MOVL new+12(FP), CX
	LOCK
	CMPXCHGL CX, 0(BX)
	SETEQ ret+16(FP)
	RET

/* func Pause() */
TEXT main·Pause(SB), NOSPLIT, $0-0
	PAUSE
	RET
