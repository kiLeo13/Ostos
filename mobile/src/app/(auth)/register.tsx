import { createUser } from '@ostos/api-contract';
import { Link, useRouter } from 'expo-router';
import { useState } from 'react';
import { StyleSheet } from 'react-native';

import { Button, ErrorNotice, Field, Screen, Title } from '@/components/ui';
import { ApiError, requireOnline, unwrap } from '@/lib/api';
import { useNetwork } from '@/providers/network-provider';
import { colors } from '@/theme';

export default function RegisterScreen() {
  const router = useRouter();
  const isOnline = useNetwork();
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<unknown>();
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    setSubmitting(true);
    setError(undefined);
    try {
      requireOnline(isOnline);
      unwrap(await createUser({ body: { email: email.trim(), username: username.trim(), display_name: displayName.trim(), password } }));
      router.replace({ pathname: '/verify-email', params: { email: email.trim() } });
    } catch (caught) {
      setError(caught);
    } finally {
      setSubmitting(false);
    }
  }

  const fields = error instanceof ApiError ? error.fields : undefined;
  return (
    <Screen>
      <Title subtitle="We will send a confirmation link before your first sign-in.">Create account</Title>
      <ErrorNotice error={error} />
      <Field label="Email" value={email} onChangeText={setEmail} autoCapitalize="none" keyboardType="email-address" error={fields?.email} />
      <Field label="Username" value={username} onChangeText={setUsername} autoCapitalize="none" autoCorrect={false} error={fields?.username} />
      <Field label="Display name" value={displayName} onChangeText={setDisplayName} error={fields?.display_name} />
      <Field label="Password" value={password} onChangeText={setPassword} secureTextEntry error={fields?.password} />
      <Button disabled={submitting} onPress={() => void submit()}>{submitting ? 'Creating…' : 'Create account'}</Button>
      <Link href="/login" style={styles.link}>Already registered? Sign in</Link>
    </Screen>
  );
}

const styles = StyleSheet.create({ link: { color: colors.primary, fontWeight: '700' } });
