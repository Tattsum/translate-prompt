export interface WebEnv {
  PAGES_HOST: string
}

const defaultPagesHost = 'translate-prompt.pages.dev'

/** Build a request targeting Cloudflare Pages (Host header must match the Pages hostname). */
export function buildPagesRequest(request: Request, pagesHost: string): Request {
  const url = new URL(request.url)
  const target = new URL(url.pathname + url.search, `https://${pagesHost}`)

  const headers = new Headers(request.headers)
  headers.set('Host', pagesHost)
  headers.set('X-Forwarded-Host', url.host)

  return new Request(target.toString(), {
    method: request.method,
    headers,
    body: request.body,
    redirect: 'manual',
  })
}

export default {
  async fetch(request: Request, env: WebEnv): Promise<Response> {
    const pagesHost = env.PAGES_HOST || defaultPagesHost
    return fetch(buildPagesRequest(request, pagesHost))
  },
} satisfies ExportedHandler<WebEnv>
