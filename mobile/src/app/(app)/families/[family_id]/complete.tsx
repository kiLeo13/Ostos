import type { ItemActionOverride } from '@ostos/api-contract';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { Button, Card, ErrorNotice, Field, Loading, Screen, Title } from '@/components/ui';
import { useFamilies } from '@/features/families/queries';
import { useCompleteList, useShoppingList } from '@/features/shopping/queries';
import { colors, spacing } from '@/theme';

export default function CompleteListScreen() {
  const { family_id: familyId } = useLocalSearchParams<{ family_id: string }>();
  const router = useRouter();
  const families = useFamilies();
  const family = families.data?.find((entry) => entry.id === familyId);
  const list = useShoppingList(familyId, family);
  const complete = useCompleteList(familyId);
  const [supermarketName, setSupermarketName] = useState('');
  const [defaultAction, setDefaultAction] = useState<'keep' | 'discard'>('keep');
  const [overrides, setOverrides] = useState<Record<string, 'keep' | 'discard'>>({});

  if (list.isLoading && !list.items.length) return <Loading />;
  const unchecked = list.items.filter((item) => !item.is_checked);
  const checked = list.items.filter((item) => item.is_checked);

  async function submit() {
    const unchecked_item_overrides: ItemActionOverride[] = Object.entries(overrides).map(([item_id, action]) => ({ item_id, action }));
    await complete.mutateAsync({
      ...(supermarketName.trim() ? { supermarket_name: supermarketName.trim() } : {}),
      default_unchecked_action: defaultAction,
      ...(unchecked_item_overrides.length ? { unchecked_item_overrides } : {}),
    });
    router.replace({ pathname: '/families/[family_id]', params: { family_id: familyId } });
  }

  return (
    <Screen>
      <Title subtitle={`${checked.length} purchased · ${unchecked.length} unchecked`}>Complete purchase</Title>
      <ErrorNotice error={complete.error} />
      <Field label="Supermarket (optional)" value={supermarketName} onChangeText={setSupermarketName} maxLength={120} />
      <Text style={styles.section}>Default for unchecked items</Text>
      <View style={styles.actions}>
        {(['keep', 'discard'] as const).map((action) => (
          <Pressable key={action} onPress={() => setDefaultAction(action)}>
            <Text style={[styles.choice, defaultAction === action && styles.selected]}>{action === 'keep' ? 'Keep for next trip' : 'Discard'}</Text>
          </Pressable>
        ))}
      </View>
      {unchecked.map((item) => {
        const action = overrides[item.id] ?? defaultAction;
        return (
          <Card key={item.id}>
            <Text style={styles.item}>{item.name}</Text>
            <View style={styles.actions}>
              <Pressable onPress={() => setOverrides((current) => ({ ...current, [item.id]: 'keep' }))}><Text style={[styles.choice, action === 'keep' && styles.selected]}>Keep</Text></Pressable>
              <Pressable onPress={() => setOverrides((current) => ({ ...current, [item.id]: 'discard' }))}><Text style={[styles.choice, action === 'discard' && styles.selected]}>Discard</Text></Pressable>
            </View>
          </Card>
        );
      })}
      <Button disabled={complete.isPending} onPress={() => void submit()}>{complete.isPending ? 'Completing…' : 'Confirm completion'}</Button>
    </Screen>
  );
}

const styles = StyleSheet.create({
  section: { color: colors.text, fontSize: 17, fontWeight: '800' },
  actions: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
  choice: { borderColor: colors.border, borderRadius: 20, borderWidth: 1, color: colors.text, paddingHorizontal: 12, paddingVertical: 8 },
  selected: { backgroundColor: colors.primary, color: '#fff' }, item: { fontSize: 16, fontWeight: '700' },
});
