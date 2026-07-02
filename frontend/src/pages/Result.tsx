import { useState } from 'react'
import { Nav } from './Input'
import { useApp } from '../context/AppContext'
import { optimize } from '../api/client'

export function ResultPage() {
  const { prompt, config, answers, optimizeResult, setOptimizeResult } = useApp()
  const [tab, setTab] = useState<'prompt' | 'task' | 'mdc'>('prompt')
  const [loading, setLoading] = useState(false)

  const result = optimizeResult
  const isCursor = config.target_profile === 'cursor'
  const mdcSuggestions = result?.artifacts?.cursor_mdc_suggestions ?? []

  async function runOptimize() {
    setLoading(true)
    try {
      const r = await optimize(prompt, config, answers)
      setOptimizeResult(r)
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Nav />
      <main>
        <h1>最適化結果</h1>
        {!result ? (
          <div className="card">
            <p>まだ最適化結果がありません。</p>
            <button onClick={runOptimize} disabled={loading || !prompt.trim()}>
              {loading ? '最適化中...' : '最適化を実行'}
            </button>
          </div>
        ) : (
          <>
            <div className="card">
              <p>
                {result.report.input_tokens} → {result.report.output_tokens} tokens (
                {result.report.reduction_percent.toFixed(1)}% 削減)
              </p>
              <ul>
                {result.report.applied_rules.map((r) => (
                  <li key={r.id}>
                    <a href={r.source_url} target="_blank" rel="noreferrer">
                      {r.id}
                    </a>
                  </li>
                ))}
              </ul>
            </div>

            {isCursor && (
              <div className="tabs">
                <button
                  className={tab === 'task' ? 'active' : ''}
                  onClick={() => setTab('task')}
                >
                  Task
                </button>
                <button
                  className={tab === 'mdc' ? 'active' : ''}
                  onClick={() => setTab('mdc')}
                >
                  .mdc 案
                </button>
                <button
                  className={tab === 'prompt' ? 'active' : ''}
                  onClick={() => setTab('prompt')}
                >
                  全文
                </button>
              </div>
            )}

            <div className="card">
              {tab === 'mdc' && isCursor ? (
                mdcSuggestions.length > 0 ? (
                  mdcSuggestions.map((m) => (
                    <div key={m.filename} style={{ marginBottom: '1rem' }}>
                      <h3>{m.filename}</h3>
                      <pre>{m.content}</pre>
                    </div>
                  ))
                ) : (
                  <p>.mdc 推奨案はありません（Rules セクションが不足している可能性があります）</p>
                )
              ) : (
                <pre>{result.optimized_prompt}</pre>
              )}
            </div>
          </>
        )}
      </main>
    </>
  )
}
