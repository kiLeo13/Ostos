import { createContext, type PropsWithChildren, useContext, useMemo, useState } from 'react';

import { useNetwork } from './network-provider';

type SyncStatus = {
  isOffline: boolean;
  lastCheckedAt?: string;
  reportApiReachability(reachable: boolean, lastCheckedAt?: string): void;
};

const SyncStatusContext = createContext<SyncStatus | null>(null);

export function SyncStatusProvider({ children }: PropsWithChildren) {
  const networkOnline = useNetwork();
  const [apiReachable, setApiReachable] = useState(true);
  const [lastCheckedAt, setLastCheckedAt] = useState<string>();
  const value = useMemo<SyncStatus>(
    () => ({
      isOffline: !networkOnline || !apiReachable,
      ...(lastCheckedAt ? { lastCheckedAt } : {}),
      reportApiReachability(reachable, checkedAt) {
        setApiReachable(reachable);
        if (checkedAt) setLastCheckedAt(checkedAt);
      },
    }),
    [apiReachable, lastCheckedAt, networkOnline],
  );
  return <SyncStatusContext.Provider value={value}>{children}</SyncStatusContext.Provider>;
}

export function useSyncStatus(): SyncStatus {
  const value = useContext(SyncStatusContext);
  if (!value) throw new Error('useSyncStatus must be used inside SyncStatusProvider');
  return value;
}
