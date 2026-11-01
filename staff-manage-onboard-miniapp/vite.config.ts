/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3001,
    host: true,
    proxy: {
      '/api/v1/onboarding': {
        target: 'http://localhost:8082',
        changeOrigin: true,
      },
      '/healthz': {
        target: 'http://localhost:8082',
        changeOrigin: true,
      },
      '/api/v1/kyc': {
        target: 'http://localhost:8085',
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
