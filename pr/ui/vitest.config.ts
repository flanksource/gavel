import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

const here = dirname(fileURLToPath(import.meta.url));

// Component tests run against real React in jsdom so the settings prompt row and
// its nested spec-editor dialog are exercised as the app renders them.
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@flanksource/gavel/testrunner/hooks': resolve(here, '../../testrunner/ui/src/hooks.ts'),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.test.{ts,tsx}'],
  },
});
