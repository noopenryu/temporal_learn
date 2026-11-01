/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api': {
        target: process.env.VITE_CLIENT_API_URL || 'http://localhost:8083',
        changeOrigin: true,
      },
      '/healthz': {
        target: process.env.VITE_CLIENT_API_URL || 'http://localhost:8083',
        changeOrigin: true,
      },
      '/loki': {
        target: process.env.VITE_LOKI_URL || 'http://localhost:3100',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/loki/, ''),
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },
});
