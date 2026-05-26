export function isDesktopRuntime(): boolean {
  return window.location.protocol === 'file:'
}

export function resolveApiBase(prefix: '/api' | '/config' | '/simulation-api'): string {
  if (!isDesktopRuntime()) {
    return prefix
  }

  if (prefix === '/api') {
    return 'http://127.0.0.1:8080'
  }
  if (prefix === '/config') {
    return 'http://127.0.0.1:8081'
  }
  return 'http://127.0.0.1:8082'
}
