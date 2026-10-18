import { Link, useRouter } from 'expo-router';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { Button, Card, Empty, ErrorNotice, Loading, Screen, Title } from '@/components/ui';
import { useFamilies } from '@/features/families/queries';
import { privateCache } from '@/lib/cache';
import { colors, spacing } from '@/theme';

export default function FamiliesScreen() {
  const router = useRouter();
  const families = useFamilies();

  async function open(familyId: string) {
    await privateCache.writeActiveFamilyId(familyId);
    router.push({ pathname: '/families/[family_id]', params: { family_id: familyId } });
  }

  if (families.isLoading) return <Loading label="Loading Families…" />;
  return (
    <Screen>
      <Title subtitle="Choose where this grocery list belongs.">Your Families</Title>
      <ErrorNotice error={families.error} />
      <View style={styles.actions}>
        <Link href="/families/new" asChild><Pressable><Text style={styles.link}>Create Family</Text></Pressable></Link>
        <Link href="/invitations" asChild><Pressable><Text style={styles.link}>Received invitations</Text></Pressable></Link>
      </View>
      {families.data?.length ? families.data.map((family) => (
        <Pressable key={family.id} onPress={() => void open(family.id)}>
          <Card>
            <Text style={styles.name}>{family.name}</Text>
            <Text style={styles.role}>{family.role}</Text>
          </Card>
        </Pressable>
      )) : <Empty>Create a Family or accept an invitation to begin.</Empty>}
      <Button variant="secondary" onPress={() => void families.refetch()}>Refresh</Button>
    </Screen>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.md },
  link: { color: colors.primary, fontWeight: '700' },
  name: { color: colors.text, fontSize: 18, fontWeight: '800' },
  role: { color: colors.muted, textTransform: 'capitalize' },
});
