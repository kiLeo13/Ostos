import * as SecureStore from 'expo-secure-store';

import type { TokenPair } from '@ostos/api-contract';

const SESSION_KEY = 'ostos.auth.session.v1';
const REFRESH_SKEW_MS = 60_000;

export type StoredSession = {
  access_token: string;
  refresh_token: string;
  access_expires_at: number;
};

export interface SessionStorage {
  get(): Promise<string | null>;
  set(value: string): Promise<void>;
  remove(): Promise<void>;
}

const secureSessionStorage: SessionStorage = {
  get: () => SecureStore.getItemAsync(SESSION_KEY),
  set: (value) =>
    SecureStore.setItemAsync(SESSION_KEY, value, {
      keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
    }),
  remove: () => SecureStore.deleteItemAsync(SESSION_KEY),
};

type RefreshTokens = (refreshToken: string) => Promise<TokenPair>;
type SessionListener = (session: StoredSession | null) => void;

export class SessionManager {
  private session: StoredSession | null = null;
  private initialized = false;
  private refreshPromise: Promise<StoredSession> | null = null;
  private readonly listeners = new Set<SessionListener>();

  constructor(
    private readonly storage: SessionStorage,
    private readonly refreshTokens: RefreshTokens,
    private readonly now: () => number = Date.now,
  ) {}

  async initialize(): Promise<StoredSession | null> {
    if (this.initialized) return this.session;
    const stored = await this.storage.get();
    this.initialized = true;
    if (!stored) return null;

    try {
      const parsed = JSON.parse(stored) as Partial<StoredSession>;
      if (
        typeof parsed.access_token !== 'string' ||
        typeof parsed.refresh_token !== 'string' ||
        typeof parsed.access_expires_at !== 'number'
      ) {
        throw new Error('invalid stored session');
      }
      this.session = parsed as StoredSession;
      return this.session;
    } catch {
      await this.storage.remove();
      return null;
    }
  }

  subscribe(listener: SessionListener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  current(): StoredSession | null {
    return this.session;
  }

  async accept(tokens: TokenPair): Promise<StoredSession> {
    const session: StoredSession = {
      access_token: tokens.access_token,
      refresh_token: tokens.refresh_token,
      access_expires_at: this.now() + tokens.expires_in * 1_000,
    };
    await this.storage.set(JSON.stringify(session));
    this.session = session;
    this.initialized = true;
    this.emit();
    return session;
  }

  async accessToken(): Promise<string | undefined> {
    await this.initialize();
    if (!this.session) return undefined;
    if (this.session.access_expires_at - this.now() <= REFRESH_SKEW_MS) {
      return (await this.refresh()).access_token;
    }
    return this.session.access_token;
  }

  async refresh(): Promise<StoredSession> {
    await this.initialize();
    if (!this.session) throw new Error('No authenticated session');
    if (this.refreshPromise) return this.refreshPromise;

    this.refreshPromise = this.refreshTokens(this.session.refresh_token)
      .then((tokens) => this.accept(tokens))
      .catch(async (error: unknown) => {
        await this.clear();
        throw error;
      })
      .finally(() => {
        this.refreshPromise = null;
      });
    return this.refreshPromise;
  }

  async clear(): Promise<void> {
    await this.storage.remove();
    this.session = null;
    this.initialized = true;
    this.emit();
  }

  private emit(): void {
    for (const listener of this.listeners) listener(this.session);
  }
}

async function refreshWithFetch(refreshToken: string): Promise<TokenPair> {
  const response = await fetch(`${apiBaseUrl()}/v1/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
  if (!response.ok) throw new Error('Session refresh failed');
  return (await response.json()) as TokenPair;
}

export function apiBaseUrl(): string {
  return (process.env.EXPO_PUBLIC_API_URL ?? 'http://127.0.0.1:8080').replace(/\/$/, '');
}

export const sessionManager = new SessionManager(secureSessionStorage, refreshWithFetch);
