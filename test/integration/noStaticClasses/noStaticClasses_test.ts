import { expect } from "chai";
import {
  CounterServiceClient,
  increment,
  failingIncrement,
  streamingIncrements,
  HttpGetRequest,
} from "./service.pb";
import { b64Decode } from "./fetch.pb";

describe("test static functions", () => {
  it("unary request", async () => {
    const result = await increment(
      { counter: 199 },
      { pathPrefix: "http://localhost:8081" }
    );

    expect(result.result).to.equal(200);
  });

  it("failing unary request", async () => {
    try {
      await failingIncrement(
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
    await streamingIncrements(
      { counter: 1 },
      (resp) => response.push(resp.result),
      { pathPrefix: "http://localhost:8081" }
    );
    expect(response).to.deep.equal([2, 3, 4, 5, 6]);
  });
});

describe("test with client class", () => {
  const client = new CounterServiceClient({
    pathPrefix: "http://localhost:8081",
  });

  it("streaming request", async () => {
    const response = [] as number[];
    await client.streamingIncrements({ counter: 1 }, (resp) =>
      response.push(resp.result)
    );
    expect(response).to.deep.equal([2, 3, 4, 5, 6]);
  });

  it("binary echo", async () => {
    const message = "→ ping";
    const resp: any = await client.echoBinary({
      data: new TextEncoder().encode(message),
    });

    const bytes = b64Decode(resp["data"]);
    expect(new TextDecoder().decode(bytes)).to.equal(message);
  });

  it("http get check request", async () => {
    const req = { numToIncrease: 10 } as HttpGetRequest;
    const result = await client.httpGet(req);
    expect(result.result).to.equal(11);
  });

  it("http post body check request with nested body path", async () => {
    const result = await client.httpPostWithNestedBodyPath({
      a: 10,
      req: { b: 15 },
      c: 0,
    });
    expect(result.postResult).to.equal(25);
  });

  it("http post body check request with star in path", async () => {
    const result = await client.httpPostWithStarBodyPath({
      a: 10,
      req: { b: 15 },
      c: 23,
    });
    expect(result.postResult).to.equal(48);
  });

  it("able to communicate with external message reference without package defined", async () => {
    const result = await client.externalMessage({ content: "hello" });
    expect(result.result).to.equal("hello!!");
  });

  it("http patch request with star in path", async () => {
    const result = await client.httpPatch({ a: 10, c: 23 });
    expect(result.patchResult).to.equal(33);
  });

  it("http delete check request", async () => {
    const result = await client.httpDelete({ a: 10 });
    expect(result).to.be.empty;
  });

  it("http delete with query params", async () => {
    const result = await client.httpDeleteWithParams({
      id: 10,
      reason: "test",
    });
    expect(result.reason).to.be.equal("test");
  });

  it("http get request with url search parameters", async () => {
    const result = await client.httpGetWithURLSearchParams({
      a: 10,
      b: { b: 0 },
      c: [23, 25],
      d: { d: 12 },
    });
    expect(result.urlSearchParamsResult).to.equal(70);
  });

  it("http get request with zero value url search parameters", async () => {
    const result = await client.httpGetWithZeroValueURLSearchParams({
      a: "A",
      b: "",
      c: { c: 1, d: [1, 0, 2], e: false },
    });
    expect(result).to.deep.equal({
      a: "A",
      b: "hello",
      zeroValueMsg: { c: 2, d: [2, 1, 3], e: true },
    });
  });

  it("http get request with optional fields", async () => {
    const result = await client.httpGetWithOptionalFields({});

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
