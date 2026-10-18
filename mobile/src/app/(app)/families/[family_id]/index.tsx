import type { ShoppingListItem, Unit } from '@ostos/api-contract';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import { Pressable, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native';

import { Button, Card, ErrorNotice, Field, Title } from '@/components/ui';
import { useFamilies } from '@/features/families/queries';
import { useListMutations, useShoppingList } from '@/features/shopping/queries';
import { formatCents, parseCents } from '@/lib/money';
import { useSyncStatus } from '@/providers/sync-status-provider';
import { colors, spacing } from '@/theme';

const units: Unit[] = ['each', 'g', 'kg', 'ml', 'l'];

function ItemCard({ item, index, allIds, familyId }: { item: ShoppingListItem; index: number; allIds: string[]; familyId: string }) {
  const actions = useListMutations(familyId);
  const [name, setName] = useState(item.name);
  const [quantity, setQuantity] = useState(item.quantity);
  const [unit, setUnit] = useState(item.unit);
  const [note, setNote] = useState(item.note ?? '');
  const [purchasedQuantity, setPurchasedQuantity] = useState(item.purchased_quantity ?? '');
  const [total, setTotal] = useState(item.total_cents == null ? '' : formatCents(item.total_cents));
  const [inputError, setInputError] = useState<unknown>();

  function move(delta: number) {
    const target = index + delta;
    if (target < 0 || target >= allIds.length) return;
    const reordered = [...allIds];
    [reordered[index], reordered[target]] = [reordered[target]!, reordered[index]!];
    actions.reorder.mutate(reordered);
  }

  function savePurchase() {
    try {
      setInputError(undefined);
      actions.update.mutate({
        itemId: item.id,
        body: {
          ...(purchasedQuantity ? { purchased_quantity: purchasedQuantity } : { clear_purchased_qty: true }),
          ...(total ? { total_cents: parseCents(total) } : { clear_total_cents: true }),
        },
      });
    } catch (caught) {
      setInputError(caught);
    }
  }

  function saveDetails() {
    actions.update.mutate({
      itemId: item.id,
      body: {
        name: name.trim(),
        quantity,
        unit,
        ...(note.trim() ? { note: note.trim() } : { clear_note: true }),
      },
    });
  }

  const error = actions.update.error ?? actions.remove.error ?? actions.reorder.error;
  return (
    <Card>
      <View style={styles.itemHeader}>
        <Pressable
          accessibilityRole="checkbox"
          accessibilityState={{ checked: item.is_checked }}
          onPress={() => actions.update.mutate({ itemId: item.id, body: { is_checked: !item.is_checked } })}
          style={[styles.checkbox, item.is_checked && styles.checkboxChecked]}
        ><Text style={styles.checkText}>{item.is_checked ? '✓' : ''}</Text></Pressable>
        <View style={styles.itemBody}>
          <Text style={[styles.itemName, item.is_checked && styles.checkedName]}>{item.name}</Text>
          <Text style={styles.meta}>{item.quantity} {item.unit}{item.note ? ` · ${item.note}` : ''}</Text>
        </View>
      </View>
      <View style={styles.purchaseFields}>
        <Field label="Name" value={name} onChangeText={setName} maxLength={160} />
        <Field label="Requested quantity" value={quantity} onChangeText={setQuantity} keyboardType="decimal-pad" />
        <View style={styles.units}>{units.map((candidate) => <Pressable key={candidate} onPress={() => setUnit(candidate)}><Text style={[styles.unit, unit === candidate && styles.unitSelected]}>{candidate}</Text></Pressable>)}</View>
        <Field label="Note" value={note} onChangeText={setNote} maxLength={500} />
        <Button variant="secondary" onPress={saveDetails}>Save item details</Button>
      </View>
      {item.is_checked ? (
        <View style={styles.purchaseFields}>
          <Field label="Bought quantity" value={purchasedQuantity} onChangeText={setPurchasedQuantity} keyboardType="decimal-pad" />
          <Field label="Total (R$)" value={total} onChangeText={setTotal} keyboardType="decimal-pad" />
          <Button variant="secondary" onPress={savePurchase}>Save purchase values</Button>
        </View>
      ) : null}
      <View style={styles.inlineActions}>
        <Pressable onPress={() => move(-1)} disabled={index === 0}><Text style={styles.action}>↑</Text></Pressable>
        <Pressable onPress={() => move(1)} disabled={index === allIds.length - 1}><Text style={styles.action}>↓</Text></Pressable>
        <Pressable onPress={() => actions.remove.mutate(item.id)}><Text style={styles.delete}>Delete</Text></Pressable>
      </View>
      <ErrorNotice error={inputError ?? error} />
    </Card>
  );
}

export default function ShoppingListScreen() {
  const { family_id: familyId } = useLocalSearchParams<{ family_id: string }>();
  const router = useRouter();
  const families = useFamilies();
  const family = families.data?.find((entry) => entry.id === familyId);
  const list = useShoppingList(familyId, family);
  const displayedFamily = family ?? list.cachedFamily;
  const actions = useListMutations(familyId);
  const sync = useSyncStatus();
  const [name, setName] = useState('');
  const [quantity, setQuantity] = useState('1');
  const [unit, setUnit] = useState<Unit>('each');

  async function add() {
    await actions.add.mutateAsync({ name: name.trim(), quantity, unit });
    setName('');
    setQuantity('1');
  }

  const ids = list.items.map((item) => item.id);
  return (
    <ScrollView
      contentContainerStyle={styles.content}
      refreshControl={<RefreshControl refreshing={list.isRefetching} onRefresh={() => void list.refetch()} />}
    >
      <Title subtitle={list.showingCached ? 'Showing the last saved snapshot.' : displayedFamily?.name}>{displayedFamily?.name ?? 'Shopping list'}</Title>
      {sync.isOffline ? <Text accessibilityRole="alert" style={styles.offline}>Read-only while offline · last checked {list.lastCheckedAt ? new Date(list.lastCheckedAt).toLocaleString() : 'never'}</Text> : null}
      <ErrorNotice error={list.error ?? actions.add.error} />
      <Card>
        <Field label="Item" value={name} onChangeText={setName} maxLength={160} editable={!sync.isOffline} />
        <Field label="Quantity" value={quantity} onChangeText={setQuantity} keyboardType="decimal-pad" editable={!sync.isOffline} />
        <View style={styles.units}>{units.map((candidate) => <Pressable key={candidate} onPress={() => setUnit(candidate)} disabled={sync.isOffline}><Text style={[styles.unit, unit === candidate && styles.unitSelected]}>{candidate}</Text></Pressable>)}</View>
        <Button disabled={sync.isOffline || !name.trim() || actions.add.isPending} onPress={() => void add()}>Add item</Button>
      </Card>
      {list.items.map((item, index) => <ItemCard key={item.id} item={item} index={index} allIds={ids} familyId={familyId} />)}
      {!list.items.length ? <Text style={styles.empty}>Nothing on the list yet.</Text> : null}
      <Button disabled={sync.isOffline || !list.items.length} onPress={() => router.push({ pathname: '/families/[family_id]/complete', params: { family_id: familyId } })}>Complete purchase</Button>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  content: { backgroundColor: colors.background, flexGrow: 1, gap: spacing.md, padding: spacing.md },
  offline: { backgroundColor: '#fff4d8', borderRadius: 8, color: colors.warning, padding: spacing.sm },
  itemHeader: { alignItems: 'center', flexDirection: 'row', gap: spacing.md },
  checkbox: { alignItems: 'center', borderColor: colors.primary, borderRadius: 6, borderWidth: 2, height: 28, justifyContent: 'center', width: 28 },
  checkboxChecked: { backgroundColor: colors.primary }, checkText: { color: '#fff', fontWeight: '900' },
  itemBody: { flex: 1 }, itemName: { color: colors.text, fontSize: 17, fontWeight: '800' }, checkedName: { textDecorationLine: 'line-through' }, meta: { color: colors.muted },
  purchaseFields: { gap: spacing.sm }, inlineActions: { flexDirection: 'row', gap: spacing.lg }, action: { color: colors.primary, fontSize: 20, fontWeight: '800' }, delete: { color: colors.danger, fontWeight: '700' },
  units: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm }, unit: { borderColor: colors.border, borderRadius: 20, borderWidth: 1, paddingHorizontal: 12, paddingVertical: 6 }, unitSelected: { backgroundColor: colors.primary, color: '#fff' },
  empty: { color: colors.muted, textAlign: 'center' },
});
