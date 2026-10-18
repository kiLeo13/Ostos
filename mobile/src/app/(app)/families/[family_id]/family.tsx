import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';

import { Button, Card, Empty, ErrorNotice, Field, Loading, Screen, Title } from '@/components/ui';
import { useFamilies, useFamilyAdministration } from '@/features/families/queries';
import { colors, spacing } from '@/theme';

export default function FamilyScreen() {
  const { family_id: familyId } = useLocalSearchParams<{ family_id: string }>();
  const router = useRouter();
  const summaries = useFamilies();
  const summary = summaries.data?.find((family) => family.id === familyId);
  const canAdminister = summary?.role === 'owner' || summary?.role === 'manager';
  const admin = useFamilyAdministration(familyId, canAdminister);
  const [nameDraft, setNameDraft] = useState<string>();
  const [invitee, setInvitee] = useState('');
  const [newOwnerId, setNewOwnerId] = useState('');
  if (admin.family.isLoading || admin.members.isLoading) return <Loading />;
  const name = nameDraft ?? admin.family.data?.name ?? '';
  const isOwner = summary?.role === 'owner';
  const error = admin.family.error ?? admin.members.error ?? admin.invitations.error ?? admin.mutation.error;

  async function exit(command: { type: 'leave' | 'delete' }) {
    await admin.mutation.mutateAsync(command);
    router.replace('/families');
  }

  return (
    <Screen>
      <Title subtitle={`Your role: ${summary?.role ?? 'member'}`}>{admin.family.data?.name ?? 'Family'}</Title>
      <ErrorNotice error={error} />
      {canAdminister ? (
        <Card>
          <Text style={styles.heading}>Settings</Text>
          <Field label="Family name" value={name} onChangeText={setNameDraft} />
          <Button variant="secondary" onPress={() => admin.mutation.mutate({ type: 'rename', name })}>Save name</Button>
          <Field label="Exact username or User ID" value={invitee} onChangeText={setInvitee} autoCapitalize="none" autoCorrect={false} />
          <Button disabled={!invitee.trim()} onPress={() => admin.mutation.mutate({ type: 'invite', identifier: invitee })}>Invite as Member</Button>
        </Card>
      ) : null}
      <Text style={styles.heading}>Members</Text>
      {admin.members.data?.map((member) => (
        <Card key={`${member.role}-${member.user_id}`}>
          <Text style={styles.member}>{member.display_name} · @{member.username}</Text>
          <Text style={styles.role}>{member.role}</Text>
          {isOwner && member.role !== 'owner' && member.id ? (
            <View style={styles.actions}>
              <View style={styles.flex}><Button variant="secondary" onPress={() => admin.mutation.mutate({ type: 'role', memberId: member.id!, role: member.role === 'manager' ? 'member' : 'manager' })}>{member.role === 'manager' ? 'Demote' : 'Promote'}</Button></View>
              <View style={styles.flex}><Button variant="danger" onPress={() => admin.mutation.mutate({ type: 'remove', memberId: member.id! })}>Remove</Button></View>
            </View>
          ) : null}
        </Card>
      ))}
      {canAdminister ? (
        <>
          <Text style={styles.heading}>Pending invitations</Text>
          {admin.invitations.data?.length ? admin.invitations.data.map((invitation) => (
            <Card key={invitation.id}>
              <Text>@{invitation.username}</Text>
              <Button variant="secondary" onPress={() => admin.mutation.mutate({ type: 'cancel-invite', invitationId: invitation.id })}>Cancel invitation</Button>
            </Card>
          )) : <Empty>No pending invitations.</Empty>}
        </>
      ) : null}
      {isOwner ? (
        <Card>
          <Text style={styles.heading}>Transfer ownership</Text>
          <Field label="New Owner User ID" value={newOwnerId} onChangeText={setNewOwnerId} keyboardType="number-pad" />
          <Button disabled={!newOwnerId} onPress={() => admin.mutation.mutate({ type: 'transfer', userId: newOwnerId })}>Transfer</Button>
          <Button variant="danger" onPress={() => void exit({ type: 'delete' })}>Delete Family</Button>
        </Card>
      ) : <Button variant="danger" onPress={() => void exit({ type: 'leave' })}>Leave Family</Button>}
    </Screen>
  );
}

const styles = StyleSheet.create({
  heading: { color: colors.text, fontSize: 18, fontWeight: '900' }, member: { fontSize: 16, fontWeight: '700' }, role: { color: colors.muted, textTransform: 'capitalize' },
  actions: { flexDirection: 'row', gap: spacing.sm }, flex: { flex: 1 },
});
