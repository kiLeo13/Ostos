import {
  completeShoppingList,
  createShoppingListItem,
  deleteShoppingListItem,
  getShoppingList,
  getShoppingTrip,
  listShoppingTrips,
  reorderShoppingListItems,
  updateShoppingListItem,
  type FamilySummary,
  type ShoppingListCompleteRequest,
  type ShoppingListItemCreateRequest,
  type ShoppingListItemUpdateRequest,
} from '@ostos/api-contract';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';

import { privateCache, type FamilyListSnapshot } from '@/lib/cache';
import { requireOnline, unwrap } from '@/lib/api';
import { queryKeys } from '@/lib/query';
import { useNetwork } from '@/providers/network-provider';
import { useSyncStatus } from '@/providers/sync-status-provider';

export function useShoppingList(familyId: string, family?: FamilySummary) {
  const online = useNetwork();
  const sync = useSyncStatus();
  const [cached, setCached] = useState<FamilyListSnapshot | null>(null);
  const query = useQuery({
    queryKey: queryKeys.list(familyId),
    queryFn: async () => unwrap(await getShoppingList({ path: { family_id: familyId } })),
    enabled: online,
  });

  useEffect(() => {
    let active = true;
    void privateCache.readSnapshot(familyId).then((snapshot) => {
      if (active) setCached(snapshot);
    });
    return () => { active = false; };
  }, [familyId]);

  useEffect(() => {
    if (!query.data || !family) return;
    const snapshot: FamilyListSnapshot = {
      family,
      items: query.data.items,
      last_checked_at: new Date().toISOString(),
    };
    sync.reportApiReachability(true, snapshot.last_checked_at);
    void privateCache.writeSnapshot(snapshot);
  }, [family, query.data]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!online || query.isError) sync.reportApiReachability(false, cached?.last_checked_at);
  }, [cached?.last_checked_at, online, query.isError]); // eslint-disable-line react-hooks/exhaustive-deps

  return {
    ...query,
    items: query.data?.items ?? cached?.items ?? [],
    lastCheckedAt: query.data ? new Date().toISOString() : cached?.last_checked_at,
    showingCached: !query.data && cached !== null,
    cachedFamily: cached?.family,
  };
}

export function useListMutations(familyId: string) {
  const online = useNetwork();
  const queryClient = useQueryClient();
  const invalidate = () => queryClient.invalidateQueries({ queryKey: queryKeys.list(familyId) });
  const add = useMutation({
    mutationFn: async (body: ShoppingListItemCreateRequest) => {
      requireOnline(online);
      return unwrap(await createShoppingListItem({ path: { family_id: familyId }, body }));
    },
    onSuccess: invalidate,
  });
  const update = useMutation({
    mutationFn: async ({ itemId, body }: { itemId: string; body: ShoppingListItemUpdateRequest }) => {
      requireOnline(online);
      return unwrap(await updateShoppingListItem({ path: { family_id: familyId, item_id: itemId }, body }));
    },
    onSuccess: invalidate,
  });
  const remove = useMutation({
    mutationFn: async (itemId: string) => {
      requireOnline(online);
      return unwrap(await deleteShoppingListItem({ path: { family_id: familyId, item_id: itemId } }));
    },
    onSuccess: invalidate,
  });
  const reorder = useMutation({
    mutationFn: async (itemIds: string[]) => {
      requireOnline(online);
      return unwrap(await reorderShoppingListItems({ path: { family_id: familyId }, body: { item_ids: itemIds } }));
    },
    onSuccess: invalidate,
  });
  return { add, update, remove, reorder };
}

export function useCompleteList(familyId: string) {
  const online = useNetwork();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (body: ShoppingListCompleteRequest) => {
      requireOnline(online);
      return unwrap(await completeShoppingList({ path: { family_id: familyId }, body }));
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.list(familyId) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.trips(familyId) }),
      ]);
    },
  });
}

export function useTrips(familyId: string) {
  return useInfiniteQuery({
    queryKey: queryKeys.trips(familyId),
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam }) =>
      unwrap(await listShoppingTrips({ path: { family_id: familyId }, query: { ...(pageParam ? { cursor: pageParam } : {}), limit: 20 } })),
    getNextPageParam: (lastPage) => lastPage?.next_cursor,
  });
}

export function useTrip(familyId: string, tripId: string) {
  return useQuery({
    queryKey: queryKeys.trip(familyId, tripId),
    queryFn: async () => unwrap(await getShoppingTrip({ path: { family_id: familyId, trip_id: tripId } })),
  });
}
