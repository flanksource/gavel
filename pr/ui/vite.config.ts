import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  // Lib/IIFE builds don't auto-replace process.env.NODE_ENV, which React and
  // @tanstack/react-query reference at runtime — define it so the bundle has no
  // bare `process` reference in the browser.
  define: {
    'process.env.NODE_ENV': JSON.stringify('production'),
  },
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
      output: {
        inlineDynamicImports: true,
        // Stable CSS filename so the Go server can go:embed it.
        assetFileNames: 'prui.[ext]',
      },
    },
  },
});
