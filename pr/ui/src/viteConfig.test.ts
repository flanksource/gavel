// @vitest-environment node
import { dirname, resolve } from 'node:path';
import { existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ConfigEnv } from 'vite';
import config from '../vite.config';

vi.mock('node:fs', async importOriginal => ({
  ...await importOriginal<typeof import('node:fs')>(),
  existsSync: vi.fn(() => true),
}));

async function aliases(command: ConfigEnv['command']) {
  if (typeof config !== 'function') throw new Error('Expected a Vite config factory');
  const resolved = await config({ command, mode: command === 'serve' ? 'development' : 'production' });
  return resolved.resolve?.alias;
}

describe('Clicky UI dependency resolution', () => {
  afterEach(() => vi.mocked(existsSync).mockReturnValue(true));

  it('uses local JavaScript and generated CSS when serving the dev UI', async () => {
    const localPackage = resolve(dirname(fileURLToPath(import.meta.url)), '../../../../clicky-ui/packages/ui');
    expect(await aliases('serve')).toEqual(expect.arrayContaining([
      { find: /^@flanksource\/clicky-ui$/, replacement: resolve(localPackage, 'src/index.ts') },
      { find: '@flanksource/clicky-ui/components', replacement: resolve(localPackage, 'src/components.ts') },
      { find: '@flanksource/clicky-ui/styles.css', replacement: resolve(localPackage, 'dist/styles.css') },
      { find: '@flanksource/clicky-ui/mdx-editor.css', replacement: resolve(localPackage, 'dist/mdx-editor.css') },
    ]));
  });

  it('keeps production builds on the published package even with a sibling checkout', async () => {
    expect(await aliases('build')).not.toEqual(expect.arrayContaining([
      expect.objectContaining({ replacement: expect.stringContaining('clicky-ui/packages/ui') }),
    ]));
  });

  it('serves the published package when the sibling checkout is absent', async () => {
    vi.mocked(existsSync).mockReturnValue(false);
    expect(await aliases('serve')).not.toEqual(expect.arrayContaining([
      expect.objectContaining({ replacement: expect.stringContaining('clicky-ui/packages/ui') }),
    ]));
  });
});
