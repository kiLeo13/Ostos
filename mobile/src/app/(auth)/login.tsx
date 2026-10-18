import { Link, useRouter } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';

import { Button, ErrorNotice, Field, Screen, Title } from '@/components/ui';
import { useAuth } from '@/providers/auth-provider';
import { colors, spacing } from '@/theme';

export default function LoginScreen() {
  const router = useRouter();
  const { login } = useAuth();
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<unknown>();
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    setSubmitting(true);
    setError(undefined);
    try {
      await login(identifier, password);
      router.replace('/families');
    } catch (caught) {
      setError(caught);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Screen>
      <Title subtitle="Your Family grocery list, without the group-chat archaeology.">Welcome back</Title>
      <ErrorNotice error={error} />
      <Field label="Email or username" value={identifier} onChangeText={setIdentifier} autoCapitalize="none" autoCorrect={false} />
      <Field label="Password" value={password} onChangeText={setPassword} secureTextEntry />
      <Button disabled={submitting || !identifier.trim() || !password} onPress={() => void submit()}>{submitting ? 'Signing in…' : 'Sign in'}</Button>
      <View style={styles.links}>
        <Link href="/forgot-password" style={styles.link}>Forgot password?</Link>
        <Text>New here? <Link href="/register" style={styles.link}>Create an account</Link></Text>
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({ links: { gap: spacing.md }, link: { color: colors.primary, fontWeight: '700' } });
