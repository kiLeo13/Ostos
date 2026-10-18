import { StyleSheet, Text, View } from 'react-native';

import { Button, Card, Empty, ErrorNotice, Loading, Screen, Title } from '@/components/ui';
import { useInvitationResponse, useReceivedInvitations } from '@/features/families/queries';
import { spacing } from '@/theme';

export default function InvitationsScreen() {
  const invitations = useReceivedInvitations();
  const response = useInvitationResponse();
  if (invitations.isLoading) return <Loading />;
  return (
    <Screen>
      <Title>Invitations</Title>
      <ErrorNotice error={invitations.error ?? response.error} />
      {invitations.data?.length ? invitations.data.map((invitation) => (
        <Card key={invitation.id}>
          <Text style={styles.name}>{invitation.family_name}</Text>
          <Text>Expires {new Date(invitation.expires_at).toLocaleDateString()}</Text>
          <View style={styles.actions}>
            <View style={styles.action}><Button disabled={response.isPending} onPress={() => response.mutate({ invitationId: invitation.id, response: 'accept' })}>Accept</Button></View>
            <View style={styles.action}><Button variant="secondary" disabled={response.isPending} onPress={() => response.mutate({ invitationId: invitation.id, response: 'decline' })}>Decline</Button></View>
          </View>
        </Card>
      )) : <Empty>No pending invitations.</Empty>}
    </Screen>
  );
}

const styles = StyleSheet.create({ name: { fontSize: 18, fontWeight: '800' }, actions: { flexDirection: 'row', gap: spacing.sm }, action: { flex: 1 } });
