import { expect } from "chai";
import { CounterService, HttpGetRequest } from "./service.pb";
import { b64Decode } from "./fetch.pb";

describe("test default configuration", () => {
  it("unary request", async () => {
    const result = await CounterService.Increment(
      { counter: 199 },
      { pathPrefix: "http://localhost:8081" }
    );

    expect(result.result).to.equal(200);
  });

  it("failing unary request", async () => {
    try {
      await CounterService.FailingIncrement(
        { counter: 199 },
        { pathPrefix: "http://localhost:8081" }
      );
      expect.fail("expected call to throw");
    } catch (e) {
      expect(e).to.have.property("message", "this increment does not work");
      expect(e).to.have.property("code", 14);
    }
  });

  it("streaming request", async () => {
    const response = [] as number[];
    await CounterService.StreamingIncrements(
      { counter: 1 },
      (resp) => response.push(resp.result),
      {
        pathPrefix: "http://localhost:8081",
      }
    );

    expect(response).to.deep.equal([2, 3, 4, 5, 6]);
  });

  it("binary echo", async () => {
    const message = "→ ping";

    const resp: any = await CounterService.EchoBinary(
      {
        data: new TextEncoder().encode(message),
      },
      { pathPrefix: "http://localhost:8081" }
    );

    const bytes = b64Decode(resp["data"]);
    expect(new TextDecoder().decode(bytes)).to.equal(message);
  });

  it("http get check request", async () => {
    const req = { numToIncrease: 10 } as HttpGetRequest;
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
    expect(result.postResult).to.equal(25);
  });

  it("http post body check request with star in path", async () => {
    const result = await CounterService.HTTPPostWithStarBodyPath(
      { a: 10, req: { b: 15 }, c: 23 },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.postResult).to.equal(48);
  });

  it("able to communicate with external message reference without package defined", async () => {
    const result = await CounterService.ExternalMessage(
      { content: "hello" },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.result).to.equal("hello!!");
  });

  it("http patch request with star in path", async () => {
    const result = await CounterService.HTTPPatch(
      { a: 10, c: 23 },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.patchResult).to.equal(33);
  });

  it("http delete check request", async () => {
    const result = await CounterService.HTTPDelete(
      { a: 10 },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result).to.be.empty;
  });

  it("http delete with query params", async () => {
    const result = await CounterService.HTTPDeleteWithParams(
      { id: 10, reason: "test" },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.reason).to.be.equal("test");
  });

  it("http get request with url search parameters", async () => {
    const result = await CounterService.HTTPGetWithURLSearchParams(
      { a: 10, b: { b: 0 }, c: [23, 25], d: { d: 12 } },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.urlSearchParamsResult).to.equal(70);
  });

  it("http get request with zero value url search parameters", async () => {
    const result = await CounterService.HTTPGetWithZeroValueURLSearchParams(
      { a: "A", b: "", c: { c: 1, d: [1, 0, 2], e: false } },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result).to.deep.equal({
      a: "A",
      b: "hello",
      zeroValueMsg: { c: 2, d: [2, 1, 3], e: true },
    });
  });

  it("http post body check request with path param pattern", async () => {
    const result = await CounterService.HTTPPostWithPathParamPattern(
      { a: "first/10", req: { b: 15 }, c: "second/is/23" },
      { pathPrefix: "http://localhost:8081" }
    );
    expect(result.postResult).to.equal("first/10second/is/23");
  });

  it("http get request with optional fields", async () => {
    const result = await CounterService.HTTPGetWithOptionalFields(
      {},
      { pathPrefix: "http://localhost:8081" }
    );

    expect(result).to.deep.equal({
      // all empty fields will be excluded.

      // only optional zero fields will be included
      zeroMsg: {},
      zeroOptStr: "",
      zeroOptNumber: 0,
      zeroOptMsg: {},

      // defined fields are the same as above.
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
