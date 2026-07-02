import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'
import { analyze, estimate } from './client'

vi.mock('@connectrpc/connect-web', () => ({
  createConnectTransport: vi.fn(() => ({})),
}))

vi.mock('@connectrpc/connect', () => ({
  createPromiseClient: vi.fn(() => ({
    health: vi.fn().mockResolvedValue({ status: 'ok' }),
    optimize: vi.fn(),
  })),
}))

vi.mock('urql', () => ({
  createClient: vi.fn(() => ({
    query: vi.fn().mockReturnValue({
      toPromise: vi.fn().mockResolvedValue({ data: { estimate: { tokens: 42 } } }),
    }),
    mutation: vi.fn().mockReturnValue({
      toPromise: vi.fn().mockResolvedValue({
        data: { analyze: { status: 'READY', questions: [], prompt: 'ok' } },
      }),
    }),
  })),
  fetchExchange: vi.fn(),
}))

describe('API client', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('estimate uses GraphQL', async () => {
    const res = await estimate('hello', 'cl100k_base')
    expect(res.tokens).toBe(42)
  })

  it('analyze uses GraphQL', async () => {
    const res = await analyze('x', {
      target_profile: 'codex',
      max_tokens: 100,
      tokenizer: 'cl100k_base',
    })
    expect(res.status).toBe('ready')
  })
})

describe('API base URL configuration', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('uses relative paths when VITE_API_BASE_URL is unset', async () => {
    const urql = await import('urql')
    const connectWeb = await import('@connectrpc/connect-web')
    await import('./client')

    expect(urql.createClient).toHaveBeenCalledWith(
      expect.objectContaining({ url: '/query' }),
    )
    expect(connectWeb.createConnectTransport).toHaveBeenCalledWith({ baseUrl: '/' })
  })

  it('uses VITE_API_BASE_URL when set', async () => {
    vi.stubEnv('VITE_API_BASE_URL', 'https://prompt-api.tattsum.com')

    const urql = await import('urql')
    const connectWeb = await import('@connectrpc/connect-web')
    await import('./client')

    expect(urql.createClient).toHaveBeenCalledWith(
      expect.objectContaining({ url: 'https://prompt-api.tattsum.com/query' }),
    )
    expect(connectWeb.createConnectTransport).toHaveBeenCalledWith({
      baseUrl: 'https://prompt-api.tattsum.com',
    })
  })
})
