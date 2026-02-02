package runner_test

import (
	"context"
	"testing"

	"github.com/buildbarn/bb-remote-execution/internal/mock"
	runner_pb "github.com/buildbarn/bb-remote-execution/pkg/proto/runner"
	"github.com/buildbarn/bb-remote-execution/pkg/runner"
	"github.com/buildbarn/bb-storage/pkg/testutil"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"
)

func TestShellScriptWrappingRunner(t *testing.T) {
	ctrl, ctx := gomock.WithContext(context.Background(), t)

	baseRunner := mock.NewMockRunnerServer(ctrl)
	r := runner.NewShellScriptWrappingRunner(baseRunner, "/usr/bin/bash")

	response := &runner_pb.RunResponse{
		ExitCode: 0,
	}

	t.Run("NonShellScript", func(t *testing.T) {
		// Regular executables should be passed through unchanged.
		request := &runner_pb.RunRequest{
			Arguments: []string{"gcc", "-o", "hello", "hello.c"},
			EnvironmentVariables: map[string]string{
				"PATH": "/bin:/usr/bin",
			},
		}
		baseRunner.EXPECT().Run(ctx, testutil.EqProto(t, request)).Return(response, nil)

		observedResponse, err := r.Run(ctx, request)
		require.NoError(t, err)
		testutil.RequireEqualProto(t, response, observedResponse)
	})

	t.Run("ShellScript", func(t *testing.T) {
		// Shell scripts should have the interpreter prepended.
		baseRunner.EXPECT().Run(ctx, testutil.EqProto(t, &runner_pb.RunRequest{
			Arguments: []string{"/usr/bin/bash", "test-setup.sh", "arg1", "arg2"},
			EnvironmentVariables: map[string]string{
				"PATH": "/bin:/usr/bin",
			},
		})).Return(response, nil)

		observedResponse, err := r.Run(ctx, &runner_pb.RunRequest{
			Arguments: []string{"test-setup.sh", "arg1", "arg2"},
			EnvironmentVariables: map[string]string{
				"PATH": "/bin:/usr/bin",
			},
		})
		require.NoError(t, err)
		testutil.RequireEqualProto(t, response, observedResponse)
	})

	t.Run("ShellScriptWithPath", func(t *testing.T) {
		// Shell scripts with paths should also be wrapped.
		baseRunner.EXPECT().Run(ctx, testutil.EqProto(t, &runner_pb.RunRequest{
			Arguments: []string{"/usr/bin/bash", "external/bazel_tools/tools/test/test-setup.sh"},
			EnvironmentVariables: map[string]string{
				"PATH": "/bin:/usr/bin",
			},
		})).Return(response, nil)

		observedResponse, err := r.Run(ctx, &runner_pb.RunRequest{
			Arguments: []string{"external/bazel_tools/tools/test/test-setup.sh"},
			EnvironmentVariables: map[string]string{
				"PATH": "/bin:/usr/bin",
			},
		})
		require.NoError(t, err)
		testutil.RequireEqualProto(t, response, observedResponse)
	})

	t.Run("ShellScriptUpperCase", func(t *testing.T) {
		// The .SH extension should also be recognized (case insensitive).
		baseRunner.EXPECT().Run(ctx, testutil.EqProto(t, &runner_pb.RunRequest{
			Arguments: []string{"/usr/bin/bash", "script.SH"},
		})).Return(response, nil)

		observedResponse, err := r.Run(ctx, &runner_pb.RunRequest{
			Arguments: []string{"script.SH"},
		})
		require.NoError(t, err)
		testutil.RequireEqualProto(t, response, observedResponse)
	})

	t.Run("EmptyArguments", func(t *testing.T) {
		// Empty arguments should be passed through unchanged.
		request := &runner_pb.RunRequest{
			Arguments: []string{},
		}
		baseRunner.EXPECT().Run(ctx, testutil.EqProto(t, request)).Return(response, nil)

		observedResponse, err := r.Run(ctx, request)
		require.NoError(t, err)
		testutil.RequireEqualProto(t, response, observedResponse)
	})
}
