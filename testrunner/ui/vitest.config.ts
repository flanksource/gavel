import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

// Tests run against real React (jsdom) so the published library surface — the
// useTestRun hook and components — is exercised exactly as external consumers
// see it. Pure-logic tests (utils, routes) are React-agnostic and pass here too.
export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.test.{ts,tsx}'],
  },
});
