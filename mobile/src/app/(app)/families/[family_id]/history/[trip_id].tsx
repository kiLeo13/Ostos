import { useLocalSearchParams } from 'expo-router';
import { StyleSheet, Text } from 'react-native';

import { Card, ErrorNotice, Loading, Screen, Title } from '@/components/ui';
import { useTrip } from '@/features/shopping/queries';
import { formatCents } from '@/lib/money';
import { colors } from '@/theme';

export default function TripDetailScreen() {
  const { family_id: familyId, trip_id: tripId } = useLocalSearchParams<{ family_id: string; trip_id: string }>();
  const trip = useTrip(familyId, tripId);
  if (trip.isLoading) return <Loading />;
  if (!trip.data) return <Screen><ErrorNotice error={trip.error ?? new Error('Purchase not found')} /></Screen>;
  const total = trip.data.items.reduce((sum, item) => sum + BigInt(item.total_cents ?? 0), 0n);
  return (
    <Screen>
      <Title subtitle={new Date(trip.data.completed_at).toLocaleString()}>{trip.data.supermarket_name ?? 'Purchase'}</Title>
      {trip.data.items.map((item) => (
        <Card key={item.id}>
          <Text style={styles.name}>{item.name}</Text>
          <Text>{item.purchased_quantity ?? item.quantity} {item.unit}</Text>
          <Text>{item.total_cents == null ? 'Price not recorded' : `R$ ${formatCents(item.total_cents)}`}</Text>
        </Card>
      ))}
      <Text style={styles.total}>Recorded total: R$ {formatCents(total)}</Text>
    </Screen>
  );
}

const styles = StyleSheet.create({ name: { color: colors.text, fontSize: 17, fontWeight: '800' }, total: { color: colors.text, fontSize: 18, fontWeight: '900', textAlign: 'right' } });
