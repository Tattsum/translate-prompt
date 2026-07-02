import { useState } from 'react'
import { useApp } from '../context/AppContext'
import { optimize } from '../api/client'
import { Page } from '../components/Layout'

export function ResultPage() {
  const { prompt, config, answers, optimizeResult, setOptimizeResult } = useApp()
  const [tab, setTab] = useState<'prompt' | 'task' | 'mdc'>('prompt')
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)

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

  async function copyText(text: string) {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Page title="最適化結果" description="トークン削減率と最適化後のプロンプトを確認できます。">
      {!result ? (
        <div className="card">
          <div className="empty-state">
            <div className="empty-state__icon" aria-hidden="true">
              ◇
            </div>
            <p className="empty-state__text">まだ最適化結果がありません。</p>
            <button onClick={runOptimize} disabled={loading || !prompt.trim()}>
              {loading && <span className="spinner" aria-hidden="true" />}
              {loading ? '最適化中...' : '最適化を実行'}
            </button>
          </div>
        </div>
      ) : (
        <>
          <div className="card">
            <div className="stats-grid">
              <div className="stat">
                <span className="stat__label">入力</span>
                <span className="stat__value">{result.report.input_tokens.toLocaleString()}</span>
              </div>
              <div className="stat">
                <span className="stat__label">出力</span>
                <span className="stat__value">{result.report.output_tokens.toLocaleString()}</span>
              </div>
              <div className="stat">
                <span className="stat__label">削減率</span>
                <span className="stat__value stat__value--success">
                  {result.report.reduction_percent.toFixed(1)}%
                </span>
              </div>
            </div>
            {result.report.applied_rules.length > 0 && (
              <>
                <p className="field__label" style={{ marginBottom: '0.5rem' }}>
                  適用ルール
                </p>
                <ul className="rule-list">
                  {result.report.applied_rules.map((r) => (
                    <li key={r.id}>
                      <a href={r.source_url} target="_blank" rel="noreferrer">
                        {r.id}
                      </a>
                    </li>
                  ))}
                </ul>
              </>
            )}
          </div>

          {isCursor && (
            <div className="tabs" role="tablist" aria-label="表示形式">
              <button
                role="tab"
                aria-selected={tab === 'task'}
                className={tab === 'task' ? 'active' : ''}
                onClick={() => setTab('task')}
              >
                Task
              </button>
              <button
                role="tab"
                aria-selected={tab === 'mdc'}
                className={tab === 'mdc' ? 'active' : ''}
                onClick={() => setTab('mdc')}
              >
                .mdc 案
              </button>
              <button
                role="tab"
                aria-selected={tab === 'prompt'}
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
                  <div key={m.filename} style={{ marginBottom: '1.5rem' }}>
                    <div className="code-block__header">
                      <h3 className="code-block__title">{m.filename}</h3>
                      <button
                        type="button"
                        className="secondary"
                        onClick={() => copyText(m.content)}
                      >
                        {copied ? 'コピー済み' : 'コピー'}
                      </button>
                    </div>
                    <pre className="code-block">{m.content}</pre>
                  </div>
                ))
              ) : (
                <p className="field__hint">
                  .mdc 推奨案はありません（Rules セクションが不足している可能性があります）
                </p>
              )
            ) : (
              <>
                <div className="code-block__header">
                  <p className="field__label" style={{ margin: 0 }}>
                    最適化プロンプト
                  </p>
                  <button
                    type="button"
                    className="secondary"
                    onClick={() => copyText(result.optimized_prompt)}
                  >
                    {copied ? 'コピー済み' : 'コピー'}
                  </button>
                </div>
                <pre className="code-block">{result.optimized_prompt}</pre>
              </>
            )}
          </div>
        </>
      )}
    </Page>
  )
}
