const apiScheme = import.meta.env.VITE_API_SCHEME || 'http'
const apiHost = import.meta.env.VITE_API_HOST || 'localhost'
const apiPort = import.meta.env.VITE_API_PORT || '9000'
const apiBasePath = import.meta.env.VITE_API_BASE_PATH || '/api/v1'
const explicitBaseUrl = import.meta.env.VITE_API_BASE_URL
const useRelativeApi = import.meta.env.DEV || import.meta.env.VITE_USE_RELATIVE_API === 'true'

const apiOrigin = `${apiScheme}://${apiHost}:${apiPort}`

export const apiConfig = {
  apiOrigin,
  apiBasePath,
  apiBaseUrl: explicitBaseUrl || (useRelativeApi ? apiBasePath : `${apiOrigin}${apiBasePath}`),
  docsUrl: useRelativeApi ? '/swagger/index.html' : `${apiOrigin}/swagger/index.html`,
}
