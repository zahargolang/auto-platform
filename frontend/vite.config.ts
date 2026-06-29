import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// API-gateway (nginx Ingress) поднят отдельно — локально через minikube,
// прописанный в /etc/hosts хост см. helm/kolesa-platform/values-minikube.yaml
// (ingress.host: kolesa). Проксируем /api и WS сюда, чтобы в браузере все
// запросы были same-origin и не упирались в CORS, которого на gateway нет.
const API_PROXY_TARGET = process.env.VITE_API_PROXY_TARGET ?? 'http://kolesa'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': {
        target: API_PROXY_TARGET,
        changeOrigin: true,
        ws: true,
      },
    },
  },
})
