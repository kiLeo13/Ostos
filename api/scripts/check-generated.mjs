import { spawnSync } from 'node:child_process';
import { mkdtempSync, readdirSync, readFileSync, rmSync, statSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, relative } from 'node:path';

const temporaryRoot = mkdtempSync(join(tmpdir(), 'ostos-openapi-'));
const temporaryOutput = join(temporaryRoot, 'generated');

try {
  const result = spawnSync(process.execPath, ['./node_modules/@hey-api/openapi-ts/bin/run.js'], {
    cwd: process.cwd(),
    env: { ...process.env, OSTOS_OPENAPI_OUTPUT: temporaryOutput },
    encoding: 'utf8',
  });
  if (result.status !== 0) {
    process.stderr.write(result.stderr || result.stdout);
    process.exit(result.status ?? 1);
  }

  const expected = snapshot(join(process.cwd(), 'generated'));
  const actual = snapshot(temporaryOutput);
  if (JSON.stringify(expected) !== JSON.stringify(actual)) {
    process.stderr.write('Generated API client is stale. Run npm run generate.\n');
    process.exit(1);
  }
} finally {
  rmSync(temporaryRoot, { recursive: true, force: true });
}

function snapshot(root) {
  const result = {};
  walk(root, root, result);
  return result;
}

function walk(root, current, result) {
  for (const entry of readdirSync(current).sort()) {
    const path = join(current, entry);
    if (statSync(path).isDirectory()) {
      walk(root, path, result);
    } else {
      result[relative(root, path).replaceAll('\\', '/')] = readFileSync(path, 'utf8');
    }
  }
}
