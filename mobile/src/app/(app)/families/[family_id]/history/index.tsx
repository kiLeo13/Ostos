import { useLocalSearchParams, useRouter } from 'expo-router';
import { Pressable, StyleSheet, Text } from 'react-native';

import { Button, Card, Empty, ErrorNotice, Loading, Screen, Title } from '@/components/ui';
import { useTrips } from '@/features/shopping/queries';
import { colors } from '@/theme';

export default function HistoryScreen() {
  const { family_id: familyId } = useLocalSearchParams<{ family_id: string }>();
  const router = useRouter();
  const history = useTrips(familyId);
  if (history.isLoading) return <Loading />;
  const trips = history.data?.pages.flatMap((page) => page?.trips ?? []) ?? [];
  return (
    <Screen>
      <Title>Purchase history</Title>
      <ErrorNotice error={history.error} />
      {trips.length ? trips.map((trip) => (
        <Pressable key={trip.id} onPress={() => router.push({ pathname: '/families/[family_id]/history/[trip_id]', params: { family_id: familyId, trip_id: trip.id } })}>
          <Card>
            <Text style={styles.market}>{trip.supermarket_name ?? 'Supermarket not recorded'}</Text>
            <Text style={styles.date}>{new Date(trip.completed_at).toLocaleString()}</Text>
          </Card>
        </Pressable>
      )) : <Empty>No completed purchases yet.</Empty>}
      {history.hasNextPage ? <Button variant="secondary" disabled={history.isFetchingNextPage} onPress={() => void history.fetchNextPage()}>Load more</Button> : null}
    </Screen>
  );
}

const styles = StyleSheet.create({ market: { color: colors.text, fontSize: 17, fontWeight: '800' }, date: { color: colors.muted } });
