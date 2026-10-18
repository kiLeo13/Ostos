import Storage from 'expo-sqlite/kv-store';

import type { FamilySummary, ShoppingListItem } from '@ostos/api-contract';

const ACTIVE_FAMILY_KEY = 'ostos.active-family.v1';
const SNAPSHOT_KEY = 'ostos.family-list-snapshot.v1';

export type FamilyListSnapshot = {
  family: FamilySummary;
  items: ShoppingListItem[];
  last_checked_at: string;
};

export interface PrivateCache {
  readActiveFamilyId(): Promise<string | null>;
  writeActiveFamilyId(familyId: string): Promise<void>;
  readSnapshot(familyId: string): Promise<FamilyListSnapshot | null>;
  writeSnapshot(snapshot: FamilyListSnapshot): Promise<void>;
  clear(): Promise<void>;
}

export const privateCache: PrivateCache = {
  readActiveFamilyId: () => Storage.getItem(ACTIVE_FAMILY_KEY),
  writeActiveFamilyId: (familyId) => Storage.setItem(ACTIVE_FAMILY_KEY, familyId),
  async readSnapshot(familyId) {
    const value = await Storage.getItem(SNAPSHOT_KEY);
    if (!value) return null;
    try {
      const snapshot = JSON.parse(value) as FamilyListSnapshot;
      return snapshot.family.id === familyId ? snapshot : null;
    } catch {
      await Storage.removeItem(SNAPSHOT_KEY);
      return null;
    }
  },
  writeSnapshot: (snapshot) => Storage.setItem(SNAPSHOT_KEY, JSON.stringify(snapshot)),
  async clear() {
    await Promise.all([Storage.removeItem(ACTIVE_FAMILY_KEY), Storage.removeItem(SNAPSHOT_KEY)]);
  },
};
