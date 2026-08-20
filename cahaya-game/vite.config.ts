import { defineConfig } from "vite";

export default defineConfig({
  base: "/cahaya/",
  server: {
    port: 5174,
    strictPort: true,
    proxy: {
      "/cahaya/ws": {
        target: "http://127.0.0.1:8888",
        ws: true,
        changeOrigin: true,
      },
      "/cahaya/api": {
        target: "http://127.0.0.1:8888",
        changeOrigin: true,
      },
      "/admin/api": {
        target: "http://127.0.0.1:8888",
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "../go-app/cahaya-dist",
    emptyOutDir: true,
    sourcemap: false,
    target: "es2022",
  },
});
