import path from 'node:path'
import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { loadEnv } from 'vite'
import { defineConfig } from 'vitest/config'

const rootDir = path.dirname(fileURLToPath(import.meta.url))

function readPort(raw: string | undefined, fallback: number) {
  const port = Number(raw)
  return Number.isFinite(port) && port > 0 ? port : fallback
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, rootDir, 'VITE_')
  const port = readPort(env.VITE_PORT, 3001)

  return {
    plugins: [react(), tailwindcss()],
    envPrefix: 'VITE_',
    resolve: {
      alias: {
        '@': path.resolve(rootDir, './src'),
      },
    },
    server: {
      port,
    },
    preview: {
      port,
    },
    test: {
      environment: 'jsdom',
    },
  }
})
