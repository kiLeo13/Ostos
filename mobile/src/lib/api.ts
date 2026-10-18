import type { ErrorEnvelope } from '@ostos/api-contract';
import { client } from '@ostos/api-contract/client';

import { apiBaseUrl, sessionManager } from './session';

export class ApiError extends Error {
  constructor(
    message: string,
    readonly code = 'REQUEST_FAILED',
    readonly status?: number,
    readonly fields?: Record<string, string>,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

type ApiResult<T> = {
  data?: T;
  error?: unknown;
  response?: Response;
};

export function unwrap<T>(result: ApiResult<T>): T {
  if (result.data !== undefined) return result.data;
  if (result.error === undefined && result.response?.ok) return result.data as T;
  const envelope = result.error as ErrorEnvelope | undefined;
  const detail = envelope?.errors;
  throw new ApiError(
    detail?.message ?? 'The request could not be completed',
    detail?.code,
    result.response?.status,
    detail?.fields,
  );
}

export function requireOnline(isOnline: boolean): void {
  if (!isOnline) {
    throw new ApiError('Connect to the internet before changing this data', 'OFFLINE');
  }
}

client.setConfig({
  baseUrl: apiBaseUrl(),
  auth: () => sessionManager.accessToken(),
  headers: { Accept: 'application/json' },
});
