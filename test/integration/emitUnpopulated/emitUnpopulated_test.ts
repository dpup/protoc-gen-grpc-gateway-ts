// Tests run against the generated client where `use_static_classes` is set to true.

import { expect } from "chai";
import { CounterService, HttpGetRequest } from "./service.pb";
import { b64Decode } from "./fetch.pb";

describe("test with emit unpopulated", () => {
  it("http get request with optional fields", async () => {
    const result = await CounterService.HTTPGetWithOptionalFields(
      {},
      { pathPrefix: "http://localhost:8081" }
    );

    // opt fields should always be undefined and zero-values should be present.
    expect(result).to.deep.equal({
      emptyStr: "",
      emptyNumber: 0,
      emptyMsg: null,
      // empty opt fields will be excluded.

      zeroStr: "",
      zeroNumber: 0,
      zeroMsg: { str: "" },
      zeroOptStr: "",
      zeroOptNumber: 0,
      zeroOptMsg: { str: "" },

      definedStr: "hello",
      definedNumber: 123,
      definedMsg: {
        str: "hello",
        optStr: "hello",
      },
      definedOptStr: "hello",
      definedOptNumber: 123,
      definedOptMsg: {
        str: "hello",
        optStr: "hello",
      },
    });
  });
});
