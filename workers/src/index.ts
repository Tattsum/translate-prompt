export interface Env {
  ORIGIN_URL: string
}

const connectPrefix = '/translate_prompt.v1.TranslatePromptService/'

/** Headers injected by Cloudflare Access that must not reach the origin. */
const accessHeaders = new Set([
  'cf-access-jwt-assertion',
  'cf-access-authenticated-user-email',
])

function isProxiedPath(pathname: string, method: string): boolean {
  if (pathname === '/query' && (method === 'POST' || method === 'OPTIONS')) {
    return true
  }
  if (pathname.startsWith(connectPrefix)) {
    return true
  }
  return false
}

function buildOriginRequest(request: Request, originURL: string): Request {
  const url = new URL(request.url)
  const target = new URL(url.pathname + url.search, originURL)

  const headers = new Headers(request.headers)
  for (const name of accessHeaders) {
    headers.delete(name)
  }

  return new Request(target.toString(), {
    method: request.method,
    headers,
    body: request.body,
    redirect: 'manual',
  })
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    if (!env.ORIGIN_URL) {
      return new Response('ORIGIN_URL is not configured', { status: 500 })
    }

    const url = new URL(request.url)
    if (!isProxiedPath(url.pathname, request.method)) {
      return new Response('Not Found', { status: 404 })
    }

    const originRequest = buildOriginRequest(request, env.ORIGIN_URL)
    // CORS is handled by the Go origin (ALLOWED_ORIGINS); proxy response as-is.
    return fetch(originRequest)
  },
} satisfies ExportedHandler<Env>
