import { defineConfig } from 'vite';

// The daemon serves /api/v1/* and /api/v1/rooms/{id}/stream (WebSocket).
// During `npm run dev`, Vite proxies to the daemon at 127.0.0.1:8787 so the
// frontend dev server can iterate without rebuilding the Go binary.
export default defineConfig({
  base: './',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
    target: 'es2020',
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8787',
        changeOrigin: true,
        ws: true,
      },
      '/healthz': 'http://127.0.0.1:8787',
    },
  },
});
