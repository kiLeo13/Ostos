import { Redirect, Stack } from 'expo-router';

import { Loading } from '@/components/ui';
import { useAuth } from '@/providers/auth-provider';

export default function AuthLayout() {
  const { status } = useAuth();
  if (status === 'loading') return <Loading />;
  if (status === 'authenticated') return <Redirect href="/families" />;
  return <Stack screenOptions={{ headerBackTitle: 'Back' }} />;
}
