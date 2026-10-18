import { login as loginRequest, logout as logoutRequest, logoutAll as logoutAllRequest } from '@ostos/api-contract';
import { createContext, type PropsWithChildren, useCallback, useContext, useEffect, useMemo, useState } from 'react';

import { privateCache } from '@/lib/cache';
import { requireOnline, unwrap } from '@/lib/api';
import { queryClient } from '@/lib/query';
import { sessionManager, type StoredSession } from '@/lib/session';
import { useNetwork } from './network-provider';

type AuthStatus = 'loading' | 'anonymous' | 'authenticated';

type AuthContextValue = {
  status: AuthStatus;
  login(identifier: string, password: string): Promise<void>;
  logout(allSessions?: boolean): Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const isOnline = useNetwork();
  const [status, setStatus] = useState<AuthStatus>('loading');

  useEffect(() => {
    let active = true;
    void sessionManager.initialize().then((session) => {
      if (active) setStatus(session ? 'authenticated' : 'anonymous');
    });
    const unsubscribe = sessionManager.subscribe((session: StoredSession | null) => {
      if (active) setStatus(session ? 'authenticated' : 'anonymous');
    });
    return () => {
      active = false;
      unsubscribe();
    };
  }, []);

  const login = useCallback(
    async (identifier: string, password: string) => {
      requireOnline(isOnline);
      const tokens = unwrap(
        await loginRequest({
          body: { identifier: identifier.trim(), password, device_name: 'Ostos mobile' },
        }),
      );
      if (!tokens) throw new Error('The API returned an empty login response');
      await sessionManager.accept(tokens);
    },
    [isOnline],
  );

  const logout = useCallback(
    async (allSessions = false) => {
      try {
        if (isOnline) {
          if (allSessions) unwrap(await logoutAllRequest());
          else unwrap(await logoutRequest());
        }
      } finally {
        await sessionManager.clear();
        await privateCache.clear();
        queryClient.clear();
      }
    },
    [isOnline],
  );

  const value = useMemo(() => ({ status, login, logout }), [login, logout, status]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const value = useContext(AuthContext);
  if (!value) throw new Error('useAuth must be used inside AuthProvider');
  return value;
}
