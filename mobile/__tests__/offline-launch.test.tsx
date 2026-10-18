import { render, screen, waitFor } from '@testing-library/react-native';

import IndexRoute from '@/app/index';
import { privateCache } from '@/lib/cache';

let mockAuthStatus: 'loading' | 'anonymous' | 'authenticated' = 'authenticated';

jest.mock('@/providers/auth-provider', () => ({
  useAuth: () => ({ status: mockAuthStatus }),
}));

jest.mock('@/lib/cache', () => ({
  privateCache: { readActiveFamilyId: jest.fn() },
}));

jest.mock('expo-router', () => {
  const { Text } = jest.requireActual('react-native') as typeof import('react-native');
  return {
    Redirect: ({ href }: { href: unknown }) => <Text>{JSON.stringify(href)}</Text>,
  };
});

describe('authenticated launch routing', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockAuthStatus = 'authenticated';
  });

  test('opens the cached active Family without waiting for the network', async () => {
    jest.mocked(privateCache.readActiveFamilyId).mockResolvedValue('241895317264384000');
    await render(<IndexRoute />);

    await waitFor(() => expect(screen.getByText(/241895317264384000/)).toBeOnTheScreen());
  });

  test('opens the chooser when no Family has been cached', async () => {
    jest.mocked(privateCache.readActiveFamilyId).mockResolvedValue(null);
    await render(<IndexRoute />);

    await waitFor(() => expect(screen.getByText('"/families"')).toBeOnTheScreen());
  });

  test('does not read private cache before authentication', async () => {
    mockAuthStatus = 'anonymous';
    await render(<IndexRoute />);

    expect(screen.getByText('"/login"')).toBeOnTheScreen();
    expect(privateCache.readActiveFamilyId).not.toHaveBeenCalled();
  });
});
