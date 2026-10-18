import { resendEmailVerification, verifyEmail } from '@ostos/api-contract';
import { Link, useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, Text } from 'react-native';

import { Button, ErrorNotice, Field, Screen, Title } from '@/components/ui';
import { requireOnline, unwrap } from '@/lib/api';
import { useNetwork } from '@/providers/network-provider';
import { colors } from '@/theme';

export default function VerifyEmailScreen() {
  const params = useLocalSearchParams<{ email?: string; token?: string }>();
  const router = useRouter();
  const isOnline = useNetwork();
  const [email, setEmail] = useState(params.email ?? '');
  const [token, setToken] = useState(params.token ?? '');
  const [message, setMessage] = useState<string>();
  const [error, setError] = useState<unknown>();

  async function verify() {
    setError(undefined);
    try {
      requireOnline(isOnline);
      unwrap(await verifyEmail({ body: { token: token.trim() } }));
      router.replace('/login');
    } catch (caught) { setError(caught); }
  }

  async function resend() {
    setError(undefined);
    try {
      requireOnline(isOnline);
      unwrap(await resendEmailVerification({ body: { email: email.trim() } }));
      setMessage('If the account is eligible, a new verification email is on its way.');
    } catch (caught) { setError(caught); }
  }

  return (
    <Screen>
      <Title subtitle="Open the link from your email, or paste its token below.">Verify email</Title>
      <ErrorNotice error={error} />
      {message ? <Text>{message}</Text> : null}
      <Field label="Verification token" value={token} onChangeText={setToken} autoCapitalize="none" />
      <Button disabled={!token.trim()} onPress={() => void verify()}>Verify</Button>
      <Field label="Email for resend" value={email} onChangeText={setEmail} autoCapitalize="none" keyboardType="email-address" />
      <Button variant="secondary" disabled={!email.trim()} onPress={() => void resend()}>Resend email</Button>
      <Link href="/login" style={styles.link}>Back to sign in</Link>
    </Screen>
  );
}

const styles = StyleSheet.create({ link: { color: colors.primary, fontWeight: '700' } });
