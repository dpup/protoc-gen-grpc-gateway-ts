package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldsWithNumbers(t *testing.T) {
	// Verifies that the proto generates the field name in the right format.

	createTestFile("fieldsWithNumbers.ts", `
import {FieldsWithNumbers} from "./names.pb"
export const newFieldsWithNumbers = (f: FieldsWithNumbers) => f.k8sField;
export const result = newFieldsWithNumbers({k8sField: 'test'});
    `)

	defer removeTestFile("fieldsWithNumbers.ts")

	result := runTsc()
	assert.Nil(t, result.err)
	assert.Equal(t, 0, result.exitCode)

	assert.Empty(t, result.stderr)
	assert.Empty(t, result.stdout)
}
