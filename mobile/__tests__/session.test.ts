import type { TokenPair } from '@ostos/api-contract';

import { SessionManager, type SessionStorage, type StoredSession } from '@/lib/session';

class MemoryStorage implements SessionStorage {
  value: string | null = null;
  removes = 0;
  async get() { return this.value; }
  async set(value: string) { this.value = value; }
  async remove() { this.value = null; this.removes++; }
}

const tokens = (suffix: string): TokenPair => ({
  access_token: `access-${suffix}`,
  refresh_token: `refresh-${suffix}`,
  token_type: 'Bearer',
  expires_in: 1800,
});

describe('SessionManager', () => {
  test('persists accepted credentials and serves a valid access token', async () => {
    const storage = new MemoryStorage();
    const refresh = jest.fn<Promise<TokenPair>, [string]>();
    const manager = new SessionManager(storage, refresh, () => 1_000);

    await manager.accept(tokens('one'));

    expect(await manager.accessToken()).toBe('access-one');
    expect(refresh).not.toHaveBeenCalled();
    expect(JSON.parse(storage.value!) as StoredSession).toEqual({
      access_token: 'access-one',
      refresh_token: 'refresh-one',
      access_expires_at: 1_801_000,
    });
  });

  test('rotates an expiring session once across concurrent callers', async () => {
    const storage = new MemoryStorage();
    storage.value = JSON.stringify({ access_token: 'old', refresh_token: 'refresh-old', access_expires_at: 50_000 });
    const refresh = jest.fn(async () => tokens('new'));
    const manager = new SessionManager(storage, refresh, () => 1_000);

    const [first, second] = await Promise.all([manager.accessToken(), manager.accessToken()]);

    expect(first).toBe('access-new');
    expect(second).toBe('access-new');
    expect(refresh).toHaveBeenCalledTimes(1);
    expect(refresh).toHaveBeenCalledWith('refresh-old');
  });

  test('clears credentials when refresh is rejected', async () => {
    const storage = new MemoryStorage();
    storage.value = JSON.stringify({ access_token: 'old', refresh_token: 'expired', access_expires_at: 0 });
    const manager = new SessionManager(storage, async () => { throw new Error('revoked'); }, () => 1_000);

    await expect(manager.accessToken()).rejects.toThrow('revoked');
    expect(manager.current()).toBeNull();
    expect(storage.value).toBeNull();
  });

  test('removes malformed persisted state instead of authenticating', async () => {
    const storage = new MemoryStorage();
    storage.value = '{not-json';
    const manager = new SessionManager(storage, async () => tokens('unused'));

    await expect(manager.initialize()).resolves.toBeNull();
    expect(storage.removes).toBe(1);
  });
});
