package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidOneOfUseCase(t *testing.T) {
	createTestFile("valid.ts", `
import {LogEntryLevel, LogService} from "./log.pb";
import {DataSource} from "./datasource/datasource.pb"
import {Environment} from "./environment.pb"

(async () => {
  const cloudSourceResult = await LogService.FetchLog({
    source: DataSource.Cloud,
    service: "cloudService"
  })

  const dataCentreSourceResult = await LogService.StreamLog({
    source: DataSource.DataCentre,
    application: "data-centre-app"
  })

  const pushLogResult = await LogService.PushLog({
    entry: {
      service: "test",
      level: LogEntryLevel.INFO,
      elapsed: 5,
      env: Environment.Production,
      message: "error message ",
      tags: ["activity1", "service1"],
      timestamp: 1592221950509.390,
      hasStackTrace: true,
      stackTraces: [{
        exception: {
          type: 'network',
          message: "timeout connecting to xyz",
        },
        lines: [{
          identifier: "A.method1",
          file: "a.java",
          line: "233",
        }],
      }],
    },
    source: DataSource.Cloud
  })
})()
    `)

	defer removeTestFile("valid.ts")

	result := runTsc()
	assert.Nil(t, result.err)
	assert.Equal(t, 0, result.exitCode)

	assert.Empty(t, result.stderr)
	assert.Empty(t, result.stdout)
}

func TestInvalidOneOfUseCase(t *testing.T) {
	createTestFile("invalid.ts", `
import {LogService} from "./log.pb";
import {DataSource} from "./datasource/datasource.pb"

(async () => {
  const cloudSourceResult = await LogService.FetchLog({
    source: DataSource.Cloud,
    service: "cloudService",
    application: "cloudApplication",
  })

})()
    `)
	defer removeTestFile("invalid.ts")

	result := runTsc()
	assert.NotNil(t, result.err)
	assert.NotEqual(t, 0, result.exitCode)

	assert.Contains(t, result.stderr,
		"Argument of type '{ source: DataSource.Cloud; service: string; application: "+
			"string; }' is not assignable to parameter of type 'FetchLogRequest'.")
}
