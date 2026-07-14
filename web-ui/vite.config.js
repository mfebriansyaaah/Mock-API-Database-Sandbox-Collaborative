import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// Vite config for Mock Sandbox Console.
// Proxies /sandbox and /hello to the local Go backend to avoid CORS in dev.
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/sandbox': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/hello': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
