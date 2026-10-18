import { Redirect } from 'expo-router';
import { useEffect, useState } from 'react';

import { Loading } from '@/components/ui';
import { privateCache } from '@/lib/cache';
import { useAuth } from '@/providers/auth-provider';

export default function IndexRoute() {
  const { status } = useAuth();
  const [activeFamilyId, setActiveFamilyId] = useState<string | null>();

  useEffect(() => {
    if (status !== 'authenticated') return;
    let active = true;
    void privateCache.readActiveFamilyId().then((familyId) => {
      if (active) setActiveFamilyId(familyId);
    });
    return () => { active = false; };
  }, [status]);

  if (status === 'loading') return <Loading label="Opening Ostos…" />;
  if (status === 'anonymous') return <Redirect href="/login" />;
  if (activeFamilyId === undefined) return <Loading label="Opening your last Family…" />;
  return activeFamilyId
    ? <Redirect href={{ pathname: '/families/[family_id]', params: { family_id: activeFamilyId } }} />
    : <Redirect href="/families" />;
}
