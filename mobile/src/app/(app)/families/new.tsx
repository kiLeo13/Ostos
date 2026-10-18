import { useRouter } from 'expo-router';
import { useState } from 'react';

import { Button, ErrorNotice, Field, Screen, Title } from '@/components/ui';
import { useCreateFamily } from '@/features/families/queries';
import { privateCache } from '@/lib/cache';

export default function NewFamilyScreen() {
  const router = useRouter();
  const create = useCreateFamily();
  const [name, setName] = useState('');

  async function submit() {
    const family = await create.mutateAsync(name);
    if (!family) throw new Error('The API returned an empty Family response');
    await privateCache.writeActiveFamilyId(family.id);
    router.replace({ pathname: '/families/[family_id]', params: { family_id: family.id } });
  }

  return (
    <Screen>
      <Title>Create Family</Title>
      <ErrorNotice error={create.error} />
      <Field label="Family name" value={name} onChangeText={setName} maxLength={100} />
      <Button disabled={!name.trim() || create.isPending} onPress={() => void submit()}>{create.isPending ? 'Creating…' : 'Create'}</Button>
    </Screen>
  );
}
