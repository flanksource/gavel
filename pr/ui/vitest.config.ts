import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

// Component tests run against real React in jsdom so the settings prompt row and
// its nested spec-editor dialog are exercised as the app renders them.
export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.test.{ts,tsx}'],
  },
});
