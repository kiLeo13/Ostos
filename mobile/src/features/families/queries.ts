import {
  acceptInvitation,
  cancelFamilyInvitation,
  createFamily,
  createFamilyInvitation,
  declineInvitation,
  deleteFamily,
  getFamily,
  leaveFamily,
  listFamilies,
  listFamilyInvitations,
  listFamilyMembers,
  listReceivedInvitations,
  removeFamilyMember,
  transferFamilyOwnership,
  updateFamily,
  updateFamilyMember,
} from '@ostos/api-contract';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { requireOnline, unwrap } from '@/lib/api';
import { queryKeys } from '@/lib/query';
import { useNetwork } from '@/providers/network-provider';

export function useFamilies() {
  return useQuery({ queryKey: queryKeys.families, queryFn: async () => unwrap(await listFamilies()) });
}

export function useCreateFamily() {
  const queryClient = useQueryClient();
  const online = useNetwork();
  return useMutation({
    mutationFn: async (name: string) => {
      requireOnline(online);
      return unwrap(await createFamily({ body: { name: name.trim() } }));
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.families }),
  });
}

export function useReceivedInvitations() {
  return useQuery({ queryKey: queryKeys.invitations, queryFn: async () => unwrap(await listReceivedInvitations()) });
}

export function useInvitationResponse() {
  const queryClient = useQueryClient();
  const online = useNetwork();
  return useMutation({
    mutationFn: async ({ invitationId, response }: { invitationId: string; response: 'accept' | 'decline' }) => {
      requireOnline(online);
      const options = { path: { invitation_id: invitationId } };
      if (response === 'accept') unwrap(await acceptInvitation(options));
      else unwrap(await declineInvitation(options));
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.invitations }),
        queryClient.invalidateQueries({ queryKey: queryKeys.families }),
      ]);
    },
  });
}

export function useFamilyAdministration(familyId: string, canAdminister: boolean) {
  const online = useNetwork();
  const queryClient = useQueryClient();
  const family = useQuery({ queryKey: queryKeys.family(familyId), queryFn: async () => unwrap(await getFamily({ path: { family_id: familyId } })) });
  const members = useQuery({ queryKey: queryKeys.members(familyId), queryFn: async () => unwrap(await listFamilyMembers({ path: { family_id: familyId } })) });
  const invitations = useQuery({ queryKey: queryKeys.familyInvitations(familyId), queryFn: async () => unwrap(await listFamilyInvitations({ path: { family_id: familyId } })), enabled: canAdminister, retry: false });
  const mutation = useMutation({
    mutationFn: async (command:
      | { type: 'rename'; name: string }
      | { type: 'invite'; identifier: string }
      | { type: 'role'; memberId: string; role: 'manager' | 'member' }
      | { type: 'remove'; memberId: string }
      | { type: 'cancel-invite'; invitationId: string }
      | { type: 'transfer'; userId: string }
      | { type: 'leave' }
      | { type: 'delete' }
    ) => {
      requireOnline(online);
      const familyPath = { family_id: familyId };
      switch (command.type) {
        case 'rename': return unwrap(await updateFamily({ path: familyPath, body: { name: command.name.trim() } }));
        case 'invite': {
          const identifier = command.identifier.trim();
          const body = /^[1-9][0-9]*$/.test(identifier) ? { user_id: identifier } : { username: identifier };
          return unwrap(await createFamilyInvitation({ path: familyPath, body }));
        }
        case 'role': return unwrap(await updateFamilyMember({ path: { ...familyPath, member_id: command.memberId }, body: { role: command.role } }));
        case 'remove': return unwrap(await removeFamilyMember({ path: { ...familyPath, member_id: command.memberId } }));
        case 'cancel-invite': return unwrap(await cancelFamilyInvitation({ path: { ...familyPath, invitation_id: command.invitationId } }));
        case 'transfer': return unwrap(await transferFamilyOwnership({ path: familyPath, body: { new_owner_id: command.userId } }));
        case 'leave': return unwrap(await leaveFamily({ path: familyPath }));
        case 'delete': return unwrap(await deleteFamily({ path: familyPath }));
      }
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.families }),
        queryClient.invalidateQueries({ queryKey: queryKeys.family(familyId) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.members(familyId) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.familyInvitations(familyId) }),
      ]);
    },
  });
  return { family, members, invitations, mutation };
}
