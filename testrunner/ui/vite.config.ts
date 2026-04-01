import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

export default defineConfig({
  plugins: [preact()],
  build: {
    lib: {
      entry: 'src/index.tsx',
      name: 'TestUI',
      formats: ['iife'],
      fileName: () => 'testui.js',
    },
    outDir: 'dist',
    minify: true,
    rollupOptions: {
      output: { inlineDynamicImports: true },
    },
  },
});
