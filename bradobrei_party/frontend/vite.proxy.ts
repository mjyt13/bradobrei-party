import type { ProxyOptions } from 'vite'

type EnvMap = Record<string, string>

export function createViteProxy(env: EnvMap): Record<string, ProxyOptions> {
  const apiScheme = env.VITE_API_SCHEME || 'http'
  const apiHost = env.VITE_API_HOST || 'localhost'
  const apiPort = env.VITE_API_PORT || '8080'
  const proxyTarget = `${apiScheme}://${apiHost}:${apiPort}`

  return {
    '/api': {
      target: proxyTarget,
      changeOrigin: true,
    },
    '/swagger': {
      target: proxyTarget,
      changeOrigin: true,
    },
    '/docs': {
      target: proxyTarget,
      changeOrigin: true,
    },
    '/health': {
      target: proxyTarget,
      changeOrigin: true,
    },
  }
}
