import { StyleSheet, Text, View } from 'react-native';

import { colors } from '@/theme';

export function OfflineStatus({ lastCheckedAt }: { lastCheckedAt?: string }) {
  const checked = lastCheckedAt ? new Date(lastCheckedAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : 'never';
  return (
    <View accessible accessibilityRole="alert" style={styles.badge}>
      <Text style={styles.text}>Offline · last checked {checked}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: { backgroundColor: '#fff4d8', borderColor: '#e0b15d', borderRadius: 8, borderWidth: 1, paddingHorizontal: 8, paddingVertical: 4 },
  text: { color: colors.warning, fontSize: 11, fontWeight: '700' },
});
