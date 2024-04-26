import { expect } from 'chai';
import camelCase from 'lodash.camelcase';
import { pathOr } from 'ramda';
import { CounterService, HttpGetRequest } from './service.pb';
import { b64Decode } from './fetch.pb';

function getFieldName(name: string) {
  const useCamelCase = pathOr(false, ['__karma__', 'config', 'useProtoNames'], window) === false;
  return useCamelCase ? camelCase(name) : name;
}

function getField(obj: { [key: string]: any }, name: string) {
  return obj[getFieldName(name)];
}

describe('test grpc-gateway-ts communication', () => {
  it('unary request', async () => {
    const result = await CounterService.Increment(
      { counter: 199 },
      { pathPrefix: 'http://localhost:8081' },
    );

    expect(result.result).to.equal(200);
  });

  it('failing unary request', async () => {
    try {
      await CounterService.FailingIncrement(
        { counter: 199 },
        { pathPrefix: 'http://localhost:8081' },
      );
      expect.fail('expected call to throw');
    } catch (e) {
      expect(e).to.have.property('message', 'this increment does not work');
      expect(e).to.have.property('code', 14);
    }
  });

  it('streaming request', async () => {
    const response = [] as number[];
    await CounterService.StreamingIncrements({ counter: 1 }, (resp) => response.push(resp.result), {
      pathPrefix: 'http://localhost:8081',
    });

    expect(response).to.deep.equal([2, 3, 4, 5, 6]);
  });

  it('binary echo', async () => {
    const message = '→ ping';

    const resp: any = await CounterService.EchoBinary(
      {
        data: new TextEncoder().encode(message),
      },
      { pathPrefix: 'http://localhost:8081' },
    );

    const bytes = b64Decode(resp['data']);
    expect(new TextDecoder().decode(bytes)).to.equal(message);
  });

  it('http get check request', async () => {
    const req = { [getFieldName('num_to_increase')]: 10 } as HttpGetRequest;
    const result = await CounterService.HTTPGet(req, { pathPrefix: 'http://localhost:8081' });
    expect(result.result).to.equal(11);
  });

  it('http post body check request with nested body path', async () => {
    const result = await CounterService.HTTPPostWithNestedBodyPath(
      { a: 10, req: { b: 15 }, c: 0 },
      { pathPrefix: 'http://localhost:8081' },
    );
    expect(getField(result, 'post_result')).to.equal(25);
  });

  it('http post body check request with star in path', async () => {
    const result = await CounterService.HTTPPostWithStarBodyPath(
      { a: 10, req: { b: 15 }, c: 23 },
      { pathPrefix: 'http://localhost:8081' },
    );
    expect(getField(result, 'post_result')).to.equal(48);
  });

  it('able to communicate with external message reference without package defined', async () => {
    const result = await CounterService.ExternalMessage(
      { content: 'hello' },
      { pathPrefix: 'http://localhost:8081' },
    );
    expect(getField(result, 'result')).to.equal('hello!!');
  });

  it('http patch request with star in path', async () => {
    const result = await CounterService.HTTPPatch(
      { a: 10, c: 23 },
      { pathPrefix: 'http://localhost:8081' },
    );
    expect(getField(result, 'patch_result')).to.equal(33);
  });

  it('http delete check request', async () => {
    const result = await CounterService.HTTPDelete(
      { a: 10 },
      { pathPrefix: 'http://localhost:8081' },
    );
    expect(result).to.be.empty;
  });

  it('http get request with url search parameters', async () => {
    const result = await CounterService.HTTPGetWithURLSearchParams(
      { a: 10, b: { b: 0 }, c: [23, 25], d: { d: 12 } },
      { pathPrefix: 'http://localhost:8081' },
    );
    expect(getField(result, 'url_search_params_result')).to.equal(70);
  });

  it('http get request with zero value url search parameters', async () => {
    const result = await CounterService.HTTPGetWithZeroValueURLSearchParams(
      { a: 'A', b: '', c: { c: 1, d: [1, 0, 2], e: false } },
      { pathPrefix: 'http://localhost:8081' },
    );
    expect(result).to.deep.equal({
      a: 'A',
      b: 'hello',
      [getFieldName('zero_value_msg')]: { c: 2, d: [2, 1, 3], e: true },
    });
  });

  it('http get request with optional fields', async () => {
    const result = await CounterService.HTTPGetWithOptionalFields(
      {},
      { pathPrefix: 'http://localhost:8081' },
    );
    const emitUnpopulated = pathOr(false, ['__karma__', 'config', 'emitUnpopulated'], window);
    if (emitUnpopulated) {
      // opt fields should always be undefined and zero-values should be present.
      expect(result).to.deep.equal({
        [getFieldName('empty_str')]: '',
        [getFieldName('empty_number')]: 0,
        [getFieldName('empty_msg')]: null,
        // empty opt fields will be excluded.

        [getFieldName('zero_str')]: '',
        [getFieldName('zero_number')]: 0,
        [getFieldName('zero_msg')]: { str: '' },
        [getFieldName('zero_opt_str')]: '',
        [getFieldName('zero_opt_number')]: 0,
        [getFieldName('zero_opt_msg')]: { str: '' },

        [getFieldName('defined_str')]: 'hello',
        [getFieldName('defined_number')]: 123,
        [getFieldName('defined_msg')]: { str: 'hello', [getFieldName('opt_str')]: 'hello' },
        [getFieldName('defined_opt_str')]: 'hello',
        [getFieldName('defined_opt_number')]: 123,
        [getFieldName('defined_opt_msg')]: { str: 'hello', [getFieldName('opt_str')]: 'hello' },
      });
    } else {
      expect(result).to.deep.equal({
        // all empty fields will be excluded.

        // only optional zero fields will be included
        [getFieldName('zero_msg')]: {},
        [getFieldName('zero_opt_str')]: '',
        [getFieldName('zero_opt_number')]: 0,
        [getFieldName('zero_opt_msg')]: {},

        // defined fields are the same as above.
        [getFieldName('defined_str')]: 'hello',
        [getFieldName('defined_number')]: 123,
        [getFieldName('defined_msg')]: { str: 'hello', [getFieldName('opt_str')]: 'hello' },
        [getFieldName('defined_opt_str')]: 'hello',
        [getFieldName('defined_opt_number')]: 123,
        [getFieldName('defined_opt_msg')]: { str: 'hello', [getFieldName('opt_str')]: 'hello' },
      });
    }
  });
});
