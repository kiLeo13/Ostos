import { resetPassword } from '@ostos/api-contract';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';

import { Button, ErrorNotice, Field, Screen, Title } from '@/components/ui';
import { requireOnline, unwrap } from '@/lib/api';
import { useNetwork } from '@/providers/network-provider';

export default function ResetPasswordScreen() {
  const { token: initialToken } = useLocalSearchParams<{ token?: string }>();
  const router = useRouter();
  const isOnline = useNetwork();
  const [token, setToken] = useState(initialToken ?? '');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<unknown>();

  async function submit() {
    setError(undefined);
    try {
      requireOnline(isOnline);
      unwrap(await resetPassword({ body: { token: token.trim(), new_password: password } }));
      router.replace('/login');
    } catch (caught) { setError(caught); }
  }

  return (
    <Screen>
      <Title>Choose a new password</Title>
      <ErrorNotice error={error} />
      <Field label="Reset token" value={token} onChangeText={setToken} autoCapitalize="none" />
      <Field label="New password" value={password} onChangeText={setPassword} secureTextEntry />
      <Button disabled={!token.trim() || !password} onPress={() => void submit()}>Reset password</Button>
    </Screen>
  );
}
