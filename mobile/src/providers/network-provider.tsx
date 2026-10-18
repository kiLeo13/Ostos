import NetInfo from '@react-native-community/netinfo';
import { onlineManager } from '@tanstack/react-query';
import { createContext, type PropsWithChildren, useContext, useEffect, useState } from 'react';

const NetworkContext = createContext(true);

function connected(isConnected: boolean | null, isInternetReachable: boolean | null): boolean {
  return isConnected !== false && isInternetReachable !== false;
}

export function NetworkProvider({ children }: PropsWithChildren) {
  const [isOnline, setOnline] = useState(true);

  useEffect(
    () =>
      NetInfo.addEventListener((state) => {
        const next = connected(state.isConnected, state.isInternetReachable);
        setOnline(next);
        onlineManager.setOnline(next);
      }),
    [],
  );

  return <NetworkContext.Provider value={isOnline}>{children}</NetworkContext.Provider>;
}

export function useNetwork(): boolean {
  return useContext(NetworkContext);
}

export { connected as deriveOnlineState };
