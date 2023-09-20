/* TODO(anton2920):
 *
 */

package main

const (
	/* NOTE(anton2920): see <sys/socket.h>. */
	AF_INET = 2
	PF_INET = AF_INET

	SOCK_STREAM = 1

	SO_REUSEPORT = 0x00000200
)

func main() {
	var lsock int = -1

	lsock = Socket(PF_INET, SOCK_STREAM, 0)
	if lsock < 0 {
		Fatal("Failed to create socket")
	}

	/* TODO(anton2920):
	 * bind(2)
	 * listen(2)
	 * accept(2)
	 */

	Print("Listening on 0.0.0.0:7070...")
}
