import { render, screen } from '@testing-library/react-native';

import { OfflineStatus } from '@/components/offline-status';

describe('OfflineStatus', () => {
  test('announces stale state and its cache timestamp accessibly', async () => {
    await render(<OfflineStatus lastCheckedAt="2026-10-18T12:30:00Z" />);
    expect(screen.getByRole('alert')).toHaveTextContent(/Offline · last checked/);
  });

  test('does not pretend an unknown snapshot was checked', async () => {
    await render(<OfflineStatus />);
    expect(screen.getByRole('alert')).toHaveTextContent(/never/);
  });
});
