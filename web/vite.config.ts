import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  base: '/ui/',
  server: {
    proxy: {
      '/pacts': 'http://localhost:8080',
      '/matrix': 'http://localhost:8080',
    },
  },
  build: {
    outDir: 'dist',
  },
})
