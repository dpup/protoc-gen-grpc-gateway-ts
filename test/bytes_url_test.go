package test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBytesURLGeneration(t *testing.T) {
	// Generate the TypeScript file first (since .pb.ts files are gitignored)
	cmd := exec.Command("protoc",
		"-I", projectRoot+"/test/testdata",
		"-I", projectRoot+"/test/integration/protos",
		"--grpc-gateway-ts_out", projectRoot+"/test/testdata",
		"--grpc-gateway-ts_opt", "logtostderr=true",
		"bytes_url.proto")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to generate TypeScript file: %s", string(output))

	// Read the generated TypeScript file
	content, err := os.ReadFile(projectRoot + "/test/testdata/bytes_url.pb.ts")
	require.NoError(t, err, "Failed to read generated TypeScript file")

	contentStr := string(content)

	// Verify TextDecoder is used for bytes fields in URLs with null safety
	t.Run("GetResource method wraps bytes field with TextDecoder and null safety", func(t *testing.T) {
		assert.Contains(t, contentStr, "req.resourceId ? new TextDecoder().decode(req.resourceId) : ''",
			"GetResource method should wrap resourceId with TextDecoder and null check")
	})

	t.Run("LoadArtefact method wraps bytes field with TextDecoder and null safety", func(t *testing.T) {
		assert.Contains(t, contentStr, "req.encodedArtefactPath ? new TextDecoder().decode(req.encodedArtefactPath) : ''",
			"LoadArtefact method should wrap encodedArtefactPath with TextDecoder and null check")
	})

	t.Run("ProcessMultiPath wraps multiple bytes fields with null safety", func(t *testing.T) {
		assert.Contains(t, contentStr, "req.firstPath ? new TextDecoder().decode(req.firstPath) : ''",
			"ProcessMultiPath should wrap firstPath with TextDecoder and null check")
		assert.Contains(t, contentStr, "req.secondPath ? new TextDecoder().decode(req.secondPath) : ''",
			"ProcessMultiPath should wrap secondPath with TextDecoder and null check")
	})

	t.Run("Non-bytes fields are not wrapped", func(t *testing.T) {
		// The parent field in LoadArtefact is a string, should not be wrapped
		lines := strings.Split(contentStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "LoadArtefact") && strings.Contains(line, "fetchRequest") {
				assert.Contains(t, line, "${req.parent}",
					"Non-bytes field 'parent' should not be wrapped with TextDecoder")
				assert.NotContains(t, line, "TextDecoder().decode(req.parent)",
					"Non-bytes field 'parent' should not be wrapped with TextDecoder")
				break
			}
		}
	})

	t.Run("JSDoc notes about UTF-8 decoding are present", func(t *testing.T) {
		assert.Contains(t, contentStr, "Note: Bytes fields in URL paths are automatically decoded to UTF-8 strings.",
			"JSDoc should mention automatic UTF-8 decoding")

		// Count occurrences - should be present for each method
		count := strings.Count(contentStr, "Note: Bytes fields in URL paths are automatically decoded to UTF-8 strings.")
		assert.Equal(t, 3, count, "JSDoc note should appear for all 3 methods")
	})

	t.Run("TypeScript types are correct", func(t *testing.T) {
		// Verify that request types have Uint8Array for bytes fields
		assert.Contains(t, contentStr, "resourceId?: Uint8Array",
			"GetResourceRequest should have resourceId as Uint8Array")
		assert.Contains(t, contentStr, "encodedArtefactPath?: Uint8Array",
			"LoadArtefactRequest should have encodedArtefactPath as Uint8Array")
		assert.Contains(t, contentStr, "firstPath?: Uint8Array",
			"MultiFieldRequest should have firstPath as Uint8Array")
		assert.Contains(t, contentStr, "secondPath?: Uint8Array",
			"MultiFieldRequest should have secondPath as Uint8Array")
	})
}

func TestBytesURLTypeScriptCompiles(t *testing.T) {
	result := runTsc()
	assert.Equal(t, 0, result.exitCode, "TypeScript compilation should succeed. Output:\n%s\n%s",
		result.stdout, result.stderr)
}
