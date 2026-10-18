import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 20_000,
      refetchOnReconnect: true,
    },
    mutations: { retry: false },
  },
});

export const queryKeys = {
  profile: ['profile'] as const,
  families: ['families'] as const,
  invitations: ['invitations'] as const,
  family: (familyId: string) => ['families', familyId] as const,
  members: (familyId: string) => ['families', familyId, 'members'] as const,
  familyInvitations: (familyId: string) => ['families', familyId, 'invitations'] as const,
  list: (familyId: string) => ['families', familyId, 'shopping-list'] as const,
  trips: (familyId: string) => ['families', familyId, 'shopping-trips'] as const,
  trip: (familyId: string, tripId: string) => ['families', familyId, 'shopping-trips', tripId] as const,
};
