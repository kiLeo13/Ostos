import { formatCents, parseCents } from '@/lib/money';

describe('money helpers', () => {
  test.each([
    ['0', 0],
    ['1', 100],
    ['1.2', 120],
    ['1,23', 123],
    ['999999.99', 99_999_999],
  ])('parses %s without binary floating point', (input, expected) => {
    expect(parseCents(input)).toBe(expected);
  });

  test.each(['1.234', '-1.00', 'R$ 1', 'one'])('rejects invalid amount %s', (input) => {
    expect(() => parseCents(input)).toThrow();
  });

  test('formats cents using integer arithmetic', () => {
    expect(formatCents(123)).toBe('1.23');
    expect(formatCents(9_007_199_254_740_991n)).toBe('90071992547409.91');
  });
});
