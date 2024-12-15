import { expect } from "chai";
import { CounterService, HttpGetRequest } from "./service.pb";
import { b64Decode } from "./fetch.pb";

describe("test with original proto names", () => {
  it("http get check request", async () => {
    const req = { num_to_increase: 10 } as HttpGetRequest;
    const result = await CounterService.HTTPGet(req, {
      pathPrefix: "http://localhost:8081",
    });
    expect(result.result).to.equal(11);
  });

  it("http post body check request with nested body path", async () => {
    const result = await CounterService.HTTPPostWithNestedBodyPath(
      { a: 10, req: { b: 15 }, c: 0 },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.post_result).to.equal(25);
  });

  it("http post body check request with star in path", async () => {
    const result = await CounterService.HTTPPostWithStarBodyPath(
      { a: 10, req: { b: 15 }, c: 23 },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.post_result).to.equal(48);
  });

  it("http patch request with star in path", async () => {
    const result = await CounterService.HTTPPatch(
      { a: 10, c: 23 },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.patch_result).to.equal(33);
  });

  it("http delete check request", async () => {
    const result = await CounterService.HTTPDelete(
      { a: 10 },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result).to.be.empty;
  });

  it("http get request with url search parameters", async () => {
    const result = await CounterService.HTTPGetWithURLSearchParams(
      { a: 10, b: { b: 0 }, c: [23, 25], d: { d: 12 } },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.url_search_params_result).to.equal(70);
  });

  it("http get request with zero value url search parameters", async () => {
    const result = await CounterService.HTTPGetWithZeroValueURLSearchParams(
      { a: "A", b: "", c: { c: 1, d: [1, 0, 2], e: false } },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result).to.deep.equal({
      a: "A",
      b: "hello",
      zero_value_msg: { c: 2, d: [2, 1, 3], e: true },
    });
  });

  it("http post body check request with path param pattern", async () => {
    const result = await CounterService.HTTPPostWithPathParamPattern(
      { a: "first/10", req: { b: 15 }, c: "second/is/23" },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.post_result).to.equal("first/10second/is/23");
  });

  it("http get request with optional fields", async () => {
    const result = await CounterService.HTTPGetWithOptionalFields(
      {},
      { pathPrefix: "http://localhost:8081" }
    );

    expect(result).to.deep.equal({
      // all empty fields will be excluded.

      // only optional zero fields will be included
      zero_msg: {},
      zero_opt_str: "",
      zero_opt_number: 0,
      zero_opt_msg: {},

      // defined fields are the same as above.
      defined_str: "hello",
      defined_number: 123,
      defined_msg: {
        str: "hello",
        opt_str: "hello",
      },
      defined_opt_str: "hello",
      defined_opt_number: 123,
      defined_opt_msg: {
        str: "hello",
        opt_str: "hello",
      },
    });
  });
});
