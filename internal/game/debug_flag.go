package game

// debugFlagArgs are the command-line spellings that enable the cheat router.
// Upstream keeps the same surface behind #ifdef _DEBUG, where "-i" and
// "+<command>" are only advertised by the debug help text (diablo.cpp:1043-1048)
// and the commands are replayed through the Lua console (diablo.cpp:896-903).
const (
	debugFlag       = "--debug"
	debugFlagSingle = "-debug"
)

// NewDebugRouterFromFlag returns an enabled router when args ask for debug
// commands, and nil otherwise: default builds keep cheats compiled out and
// every command a no-op.
func NewDebugRouterFromFlag(args []string, env DebugEnv) *DebugRouter {
	for _, arg := range args {
		if arg == debugFlag || arg == debugFlagSingle {
			return NewDebugRouter(env)
		}
	}

	return nil
}
