package runner

import (
	"context"
	"strings"

	runner_pb "github.com/buildbarn/bb-remote-execution/pkg/proto/runner"

	"google.golang.org/protobuf/proto"
)

type shellScriptWrappingRunner struct {
	runner_pb.RunnerServer
	interpreterPath string
}

// NewShellScriptWrappingRunner creates a decorator for RunnerServer that
// wraps shell script invocations with an interpreter. When argv[0] ends
// with ".sh", the interpreter path is prepended to the arguments list.
//
// This decorator is primarily intended for Windows workers where shell
// scripts cannot be executed directly. By setting interpreterPath to the
// path of a bash interpreter (e.g., "C:\Program Files\Git\bin\bash.exe"),
// shell scripts from cross-compiled builds can be executed.
func NewShellScriptWrappingRunner(base runner_pb.RunnerServer, interpreterPath string) runner_pb.RunnerServer {
	return &shellScriptWrappingRunner{
		RunnerServer:    base,
		interpreterPath: interpreterPath,
	}
}

func (r *shellScriptWrappingRunner) Run(ctx context.Context, oldRequest *runner_pb.RunRequest) (*runner_pb.RunResponse, error) {
	if len(oldRequest.Arguments) == 0 {
		return r.RunnerServer.Run(ctx, oldRequest)
	}

	// Check if argv[0] looks like a shell script.
	argv0 := oldRequest.Arguments[0]
	if !strings.HasSuffix(strings.ToLower(argv0), ".sh") {
		return r.RunnerServer.Run(ctx, oldRequest)
	}

	// Prepend the interpreter to the arguments.
	var newRequest runner_pb.RunRequest
	proto.Merge(&newRequest, oldRequest)
	newRequest.Arguments = make([]string, len(oldRequest.Arguments)+1)
	newRequest.Arguments[0] = r.interpreterPath
	copy(newRequest.Arguments[1:], oldRequest.Arguments)

	return r.RunnerServer.Run(ctx, &newRequest)
}
