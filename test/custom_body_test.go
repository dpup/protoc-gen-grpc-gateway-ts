package test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomBodyFieldNaming(t *testing.T) {
	// Generate the TypeScript file first (since .pb.ts files are gitignored)
	cmd := exec.Command("protoc",
		"-I", projectRoot+"/test/testdata",
		"-I", projectRoot+"/test/integration/protos",
		"--grpc-gateway-ts_out", projectRoot+"/test/testdata",
		"--grpc-gateway-ts_opt", "logtostderr=true",
		"custom_body.proto")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to generate TypeScript file: %s", string(output))

	// Read the generated TypeScript file
	content, err := os.ReadFile(projectRoot + "/test/testdata/custom_body.pb.ts")
	require.NoError(t, err, "Failed to read generated TypeScript file")

	contentStr := string(content)

	t.Run("UpdateUser uses camelCase field name for custom body", func(t *testing.T) {
		// Should use req["userUpdate"] not req["user_update"]
		assert.Contains(t, contentStr, `req["userUpdate"]`,
			"UpdateUser should use camelCase field name 'userUpdate'")
		assert.NotContains(t, contentStr, `req["user_update"]`,
			"UpdateUser should not use snake_case field name 'user_update'")
	})

	t.Run("CreatePost uses camelCase field name for custom body", func(t *testing.T) {
		// Should use req["postContent"] not req["post_content"]
		assert.Contains(t, contentStr, `req["postContent"]`,
			"CreatePost should use camelCase field name 'postContent'")
		assert.NotContains(t, contentStr, `req["post_content"]`,
			"CreatePost should not use snake_case field name 'post_content'")
	})

	t.Run("TypeScript types are correct", func(t *testing.T) {
		// Verify that request types have camelCase for fields
		assert.Contains(t, contentStr, "userId?: string",
			"UpdateUserRequest should have userId as string")
		assert.Contains(t, contentStr, "userUpdate?: UserUpdate",
			"UpdateUserRequest should have userUpdate as UserUpdate")
		assert.Contains(t, contentStr, "authorId?: string",
			"CreatePostRequest should have authorId as string")
		assert.Contains(t, contentStr, "postContent?: PostContent",
			"CreatePostRequest should have postContent as PostContent")
	})
}

func TestCustomBodyTypeScriptCompiles(t *testing.T) {
	result := runTsc()
	assert.Equal(t, 0, result.exitCode, "TypeScript compilation should succeed. Output:\n%s\n%s",
		result.stdout, result.stderr)
}
