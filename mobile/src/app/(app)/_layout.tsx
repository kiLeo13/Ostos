import { Redirect, Stack } from 'expo-router';

import { Loading } from '@/components/ui';
import { useAuth } from '@/providers/auth-provider';

export default function AppLayout() {
  const { status } = useAuth();
  if (status === 'loading') return <Loading />;
  if (status === 'anonymous') return <Redirect href="/login" />;
  return <Stack screenOptions={{ headerBackTitle: 'Back' }} />;
}
