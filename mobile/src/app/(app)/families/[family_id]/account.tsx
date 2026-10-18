import { changeCurrentUserPassword, deleteCurrentUser, getCurrentUser, updateCurrentUser } from '@ostos/api-contract';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Text } from 'react-native';

import { Button, Card, ErrorNotice, Field, Loading, Screen, Title } from '@/components/ui';
import { requireOnline, unwrap } from '@/lib/api';
import { queryKeys } from '@/lib/query';
import { useAuth } from '@/providers/auth-provider';
import { useNetwork } from '@/providers/network-provider';

export default function AccountScreen() {
  const router = useRouter();
  const online = useNetwork();
  const { logout } = useAuth();
  const queryClient = useQueryClient();
  const profile = useQuery({ queryKey: queryKeys.profile, queryFn: async () => unwrap(await getCurrentUser()) });
  const [usernameDraft, setUsernameDraft] = useState<string>();
  const [displayNameDraft, setDisplayNameDraft] = useState<string>();
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const mutation = useMutation({
    mutationFn: async (command: 'profile' | 'password' | 'delete') => {
      requireOnline(online);
      if (command === 'profile') return unwrap(await updateCurrentUser({ body: { username: username.trim(), display_name: displayName.trim() } }));
      if (command === 'password') return unwrap(await changeCurrentUserPassword({ body: { current_password: currentPassword, new_password: newPassword } }));
      return unwrap(await deleteCurrentUser({ body: { current_password: currentPassword } }));
    },
    onSuccess: (_, command) => {
      if (command === 'profile') void queryClient.invalidateQueries({ queryKey: queryKeys.profile });
    },
  });
  if (profile.isLoading) return <Loading />;
  const username = usernameDraft ?? profile.data?.username ?? '';
  const displayName = displayNameDraft ?? profile.data?.display_name ?? '';
  async function deleteAccount() {
    await mutation.mutateAsync('delete');
    await logout(true);
    router.replace('/login');
  }
  async function changePassword() {
    await mutation.mutateAsync('password');
    await logout(false);
    router.replace('/login');
  }
  return (
    <Screen>
      <Title>Account</Title>
      <ErrorNotice error={profile.error ?? mutation.error} />
      <Card>
        <Text>{profile.data?.email}</Text>
        <Field label="Username" value={username} onChangeText={setUsernameDraft} autoCapitalize="none" />
        <Field label="Display name" value={displayName} onChangeText={setDisplayNameDraft} />
        <Button variant="secondary" onPress={() => mutation.mutate('profile')}>Save profile</Button>
      </Card>
      <Card>
        <Field label="Current password" value={currentPassword} onChangeText={setCurrentPassword} secureTextEntry />
        <Field label="New password" value={newPassword} onChangeText={setNewPassword} secureTextEntry />
        <Button disabled={!currentPassword || !newPassword} onPress={() => void changePassword()}>Change password</Button>
      </Card>
      <Button variant="secondary" onPress={() => void logout(false)}>Log out this session</Button>
      <Button variant="secondary" onPress={() => void logout(true)}>Log out all sessions</Button>
      <Button variant="danger" disabled={!currentPassword} onPress={() => void deleteAccount()}>Delete account</Button>
    </Screen>
  );
}
