import { Tabs } from 'expo-router';
import { StyleSheet, Text, View } from 'react-native';

import { OfflineStatus } from '@/components/offline-status';
import { useSyncStatus } from '@/providers/sync-status-provider';
import { colors } from '@/theme';

function TabMark({ symbol, grocery = false }: { symbol: string; grocery?: boolean }) {
  const sync = useSyncStatus();
  return (
    <View style={styles.mark}>
      {grocery && sync.isOffline ? <View style={styles.offline}><OfflineStatus lastCheckedAt={sync.lastCheckedAt} /></View> : null}
      <Text style={styles.symbol}>{symbol}</Text>
    </View>
  );
}

export default function FamilyTabs() {
  return (
    <Tabs screenOptions={{ tabBarActiveTintColor: colors.primary, headerShown: true }}>
      <Tabs.Screen name="index" options={{ title: 'List', tabBarIcon: () => <TabMark symbol="✓" grocery /> }} />
      <Tabs.Screen name="history/index" options={{ title: 'History', tabBarIcon: () => <TabMark symbol="↶" /> }} />
      <Tabs.Screen name="family" options={{ title: 'Family', tabBarIcon: () => <TabMark symbol="⌂" /> }} />
      <Tabs.Screen name="account" options={{ title: 'Account', tabBarIcon: () => <TabMark symbol="●" /> }} />
      <Tabs.Screen name="complete" options={{ href: null, title: 'Complete purchase' }} />
      <Tabs.Screen name="history/[trip_id]" options={{ href: null, title: 'Purchase detail' }} />
    </Tabs>
  );
}

const styles = StyleSheet.create({
  mark: { alignItems: 'center', height: 26, justifyContent: 'center', minWidth: 44 },
  symbol: { color: colors.primary, fontSize: 20, fontWeight: '800' },
  offline: { left: -78, position: 'absolute', top: -20, width: 170, zIndex: 3 },
});
