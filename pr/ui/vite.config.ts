import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

export default defineConfig({
  plugins: [preact()],
  build: {
    lib: {
      entry: 'src/index.tsx',
      name: 'PRUI',
      formats: ['iife'],
      fileName: () => 'prui.js',
    },
    outDir: 'dist',
    minify: true,
    rollupOptions: {
      output: { inlineDynamicImports: true },
    },
  },
});
