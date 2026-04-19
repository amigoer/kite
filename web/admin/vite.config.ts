import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";
import pkg from "./package.json" with { type: "json" };

export default defineConfig({
  base: '/',
  plugins: [react(), tailwindcss()],
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version),
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      "/api": "http://localhost:8080",
      "/i/": "http://localhost:8080",
      "/v/": "http://localhost:8080",
      "/a/": "http://localhost:8080",
      "/f/": "http://localhost:8080",
      "/t/": "http://localhost:8080",
    },
  },
});
