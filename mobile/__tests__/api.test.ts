import { requireOnline, unwrap } from '@/lib/api';

jest.mock('@ostos/api-contract/client', () => ({ client: { setConfig: jest.fn() } }));

describe('API result handling', () => {
  test('returns successful JSON and successful empty responses', () => {
    expect(unwrap({ data: { id: '1' }, response: new Response(null, { status: 200 }) })).toEqual({ id: '1' });
    expect(unwrap<void>({ response: new Response(null, { status: 204 }) })).toBeUndefined();
  });

  test('preserves stable code, status, and all field errors', () => {
    expect(() => unwrap({
      error: { errors: { code: 'INVALID_FORM_BODY', message: 'Invalid', fields: { email: 'Bad email', username: 'Bad username' } } },
      response: new Response(null, { status: 422 }),
    })).toThrow(expect.objectContaining({
      code: 'INVALID_FORM_BODY',
      status: 422,
      fields: { email: 'Bad email', username: 'Bad username' },
    }));
  });

  test('blocks mutations before they can claim offline success', () => {
    expect(() => requireOnline(false)).toThrow(expect.objectContaining({ code: 'OFFLINE' }));
    expect(() => requireOnline(true)).not.toThrow();
  });
});
