import type { PropsWithChildren, ReactNode } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  type TextInputProps,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { ApiError } from '@/lib/api';
import { colors, spacing } from '@/theme';

export function Screen({ children, scroll = true }: PropsWithChildren<{ scroll?: boolean }>) {
  const content = <View style={styles.content}>{children}</View>;
  return (
    <SafeAreaView style={styles.safe} edges={['bottom']}>
      {scroll ? <ScrollView contentContainerStyle={styles.scroll}>{content}</ScrollView> : content}
    </SafeAreaView>
  );
}

export function Title({ children, subtitle }: PropsWithChildren<{ subtitle?: string }>) {
  return (
    <View style={styles.titleBlock}>
      <Text accessibilityRole="header" style={styles.title}>{children}</Text>
      {subtitle ? <Text style={styles.subtitle}>{subtitle}</Text> : null}
    </View>
  );
}

export function Field({ label, error, ...props }: TextInputProps & { label: string; error?: string }) {
  return (
    <View style={styles.field}>
      <Text style={styles.label}>{label}</Text>
      <TextInput
        accessibilityLabel={label}
        placeholderTextColor={colors.muted}
        style={[styles.input, error ? styles.inputError : undefined]}
        {...props}
      />
      {error ? <Text style={styles.errorText}>{error}</Text> : null}
    </View>
  );
}

export function Button({
  children,
  onPress,
  disabled = false,
  variant = 'primary',
}: PropsWithChildren<{ onPress(): void; disabled?: boolean; variant?: 'primary' | 'secondary' | 'danger' }>) {
  return (
    <Pressable
      accessibilityRole="button"
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.button,
        variant === 'secondary' && styles.buttonSecondary,
        variant === 'danger' && styles.buttonDanger,
        pressed && !disabled && styles.buttonPressed,
        disabled && styles.buttonDisabled,
      ]}
    >
      <Text style={[styles.buttonText, variant === 'secondary' && styles.buttonSecondaryText]}>{children}</Text>
    </Pressable>
  );
}

export function Card({ children }: PropsWithChildren) {
  return <View style={styles.card}>{children}</View>;
}

export function ErrorNotice({ error }: { error: unknown }) {
  if (!error) return null;
  const message = error instanceof ApiError || error instanceof Error ? error.message : 'Something went wrong';
  return <Text accessibilityRole="alert" style={styles.errorNotice}>{message}</Text>;
}

export function Empty({ children }: PropsWithChildren) {
  return <Text style={styles.empty}>{children}</Text>;
}

export function Loading({ label = 'Loading…' }: { label?: string }) {
  return <View style={styles.loading}><ActivityIndicator color={colors.primary} /><Text>{label}</Text></View>;
}

export function Row({ children, trailing }: PropsWithChildren<{ trailing?: ReactNode }>) {
  return <View style={styles.row}><View style={styles.rowMain}>{children}</View>{trailing}</View>;
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  scroll: { flexGrow: 1 },
  content: { flex: 1, padding: spacing.md, gap: spacing.md },
  titleBlock: { gap: spacing.xs, marginVertical: spacing.sm },
  title: { color: colors.text, fontSize: 30, fontWeight: '800' },
  subtitle: { color: colors.muted, fontSize: 16, lineHeight: 22 },
  field: { gap: spacing.xs },
  label: { color: colors.text, fontSize: 14, fontWeight: '700' },
  input: { backgroundColor: colors.surface, borderColor: colors.border, borderRadius: 10, borderWidth: 1, color: colors.text, fontSize: 16, paddingHorizontal: 14, paddingVertical: 12 },
  inputError: { borderColor: colors.danger },
  errorText: { color: colors.danger, fontSize: 13 },
  button: { alignItems: 'center', backgroundColor: colors.primary, borderRadius: 10, minHeight: 46, justifyContent: 'center', paddingHorizontal: spacing.md },
  buttonSecondary: { backgroundColor: colors.surface, borderColor: colors.primary, borderWidth: 1 },
  buttonDanger: { backgroundColor: colors.danger },
  buttonPressed: { opacity: 0.82 },
  buttonDisabled: { backgroundColor: colors.disabled, opacity: 0.7 },
  buttonText: { color: '#fff', fontSize: 16, fontWeight: '700' },
  buttonSecondaryText: { color: colors.primary },
  card: { backgroundColor: colors.surface, borderColor: colors.border, borderRadius: 12, borderWidth: 1, gap: spacing.sm, padding: spacing.md },
  errorNotice: { backgroundColor: '#fbeaea', borderRadius: 8, color: colors.danger, padding: 12 },
  empty: { color: colors.muted, fontSize: 16, paddingVertical: spacing.lg, textAlign: 'center' },
  loading: { alignItems: 'center', flex: 1, gap: spacing.sm, justifyContent: 'center' },
  row: { alignItems: 'center', flexDirection: 'row', gap: spacing.sm },
  rowMain: { flex: 1 },
});

export const uiStyles = styles;
