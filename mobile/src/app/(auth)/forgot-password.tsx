import { forgotPassword } from '@ostos/api-contract';
import { Link } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, Text } from 'react-native';

import { Button, ErrorNotice, Field, Screen, Title } from '@/components/ui';
import { requireOnline, unwrap } from '@/lib/api';
import { useNetwork } from '@/providers/network-provider';
import { colors } from '@/theme';

export default function ForgotPasswordScreen() {
  const isOnline = useNetwork();
  const [email, setEmail] = useState('');
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<unknown>();

  async function submit() {
    setError(undefined);
    try {
      requireOnline(isOnline);
      unwrap(await forgotPassword({ body: { email: email.trim() } }));
      setSent(true);
    } catch (caught) { setError(caught); }
  }

  return (
    <Screen>
      <Title subtitle="The response is intentionally the same whether an account exists or not.">Reset password</Title>
      <ErrorNotice error={error} />
      {sent ? <Text>Check your inbox for the reset link.</Text> : null}
      <Field label="Email" value={email} onChangeText={setEmail} autoCapitalize="none" keyboardType="email-address" />
      <Button disabled={!email.trim()} onPress={() => void submit()}>Send reset email</Button>
      <Link href="/reset-password" style={styles.link}>I already have a reset token</Link>
    </Screen>
  );
}

const styles = StyleSheet.create({ link: { color: colors.primary, fontWeight: '700' } });
