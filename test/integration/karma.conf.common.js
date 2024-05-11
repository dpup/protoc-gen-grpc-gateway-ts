module.exports = function(ci, config) {
  return {
    frameworks: ["mocha", "chai", "karma-typescript"],
    preprocessors: {
      "**/*.ts": ["karma-typescript"],
    },
    files: ["**/*_test.ts", "**/*.pb.ts"],
    reporters: ["progress", "karma-typescript"],
    port: 9876, // karma web server port
    colors: true,
    logLevel: config.LOG_INFO,
    browsers: ["ChromeHeadless"],
    autoWatch: false,
    singleRun: ci, // Karma captures browsers, runs the tests and exits
    concurrency: Infinity,
    karmaTypescriptConfig: {
      tsconfig: "./tsconfig.json",
    },
    client: {
      useProtoNames: process.env["USE_PROTO_NAMES"] === "true",
      emitUnpopulated: process.env["EMIT_UNPOPULATED"] === "true",
    },
  };
};
