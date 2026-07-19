import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // When the Go agent exposes an HTTP API, proxy /api calls here
      "/api": {
        target: "http://localhost:8081",
        changeOrigin: true,
      },
    },
  },
});
