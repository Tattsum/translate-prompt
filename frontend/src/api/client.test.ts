import { describe, expect, it, vi, beforeEach } from 'vitest'
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
