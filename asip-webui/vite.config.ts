import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const basePath = process.env.VITE_BASE_PATH || '/'
const apiProxyTarget = process.env.VITE_API_PROXY || 'http://127.0.0.1:3000'
const baseNoSlash = basePath.replace(/\/$/, '') || ''

export default defineConfig({
  base: basePath,
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
    allowedHosts: true,
    hmr: {
      clientPort: Number(process.env.VITE_HMR_CLIENT_PORT || 5173),
    },
    proxy: {
      [`${baseNoSlash}/api`]: {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(new RegExp(`^${baseNoSlash}/api`), '/api'),
      },
    },
  },
})
