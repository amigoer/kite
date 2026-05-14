/**
 * Runtime configuration discovered from the page URL.
 *
 * When the panel is served by a standalone daemon (`kite serve`) the
 * bundle is at the document root, so `basePath()` returns `""` and
 * everything looks like `/api/v1/...`.
 *
 * When served by a hub (`kite hub`) the bundle is rooted at
 * `/d/<daemon-name>/`, so `basePath()` returns `/d/<daemon-name>` and
 * every API/WS URL must be prefixed with it before hitting the network.
 *
 * The base path is captured once at module load; subsequent calls are
 * cache-free no-ops. Navigations within the SPA happen via `location.hash`
 * and don't change the pathname, so a one-shot read is safe.
 */
const cachedBasePath = (() => {
  const p = (typeof window !== 'undefined' && window.location?.pathname) || '/';
  // Strip a trailing index.html that some hosting setups leave in.
  let stripped = p.replace(/\/index\.html$/i, '');
  // Drop the trailing slash (so we can always join with `${base}${path}`).
  if (stripped.endsWith('/') && stripped !== '/') {
    stripped = stripped.slice(0, -1);
  }
  if (stripped === '' || stripped === '/') return '';
  return stripped;
})();

/** Path prefix for every API/WS URL the bundle constructs. */
export function basePath(): string {
  return cachedBasePath;
}

/** Build an API URL from a leading-slash path. */
export function apiURL(path: string): string {
  return cachedBasePath + path;
}

/** Build a WebSocket URL from a leading-slash path. */
export function wsURL(path: string): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${proto}//${window.location.host}${cachedBasePath}${path}`;
}

/**
 * If the bundle is running under a hub at /d/<name>/, return that name.
 * Returns null for direct-daemon deployments.
 */
export function daemonName(): string | null {
  const m = cachedBasePath.match(/^\/d\/([^/]+)$/);
  return m ? m[1] : null;
}
