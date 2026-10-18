import { deriveOnlineState } from '@/providers/network-provider';

describe('deriveOnlineState', () => {
  test.each([
    [false, null],
    [true, false],
    [false, false],
  ])('is offline for connected=%s reachable=%s', (connected, reachable) => {
    expect(deriveOnlineState(connected, reachable)).toBe(false);
  });

  test.each([
    [true, true],
    [true, null],
    [null, null],
  ])('does not invent an outage for connected=%s reachable=%s', (connected, reachable) => {
    expect(deriveOnlineState(connected, reachable)).toBe(true);
  });
});
