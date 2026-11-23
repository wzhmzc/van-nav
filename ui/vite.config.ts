import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// 从 CRA(react-scripts) 迁到 Vite：构建时间从 ~90s 降到 10s 级别。
// 输出目录保持 dist/，由 CI / Dockerfile 拷到后端的 public/ 再编进二进制。
export default defineConfig({
  plugins: [react()],
  base: '/',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    chunkSizeWarningLimit: 1500,
  },
  server: {
    port: 2333,
    proxy: {
      '/api': 'http://localhost:6412',
    },
  },
});
