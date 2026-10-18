export function parseCents(value: string): number {
  const normalized = value.trim().replace(',', '.');
  const match = /^(\d+)(?:\.(\d{0,2}))?$/.exec(normalized);
  if (!match?.[1]) throw new Error('Enter an amount with at most two decimal places');
  const fraction = (match[2] ?? '').padEnd(2, '0');
  const cents = BigInt(match[1]) * 100n + BigInt(fraction || '0');
  if (cents > BigInt(Number.MAX_SAFE_INTEGER)) throw new Error('Amount is too large');
  return Number(cents);
}

export function formatCents(value: number | bigint): string {
  const cents = typeof value === 'bigint' ? value : BigInt(value);
  const whole = cents / 100n;
  const fraction = (cents % 100n).toString().padStart(2, '0');
  return `${whole}.${fraction}`;
}
