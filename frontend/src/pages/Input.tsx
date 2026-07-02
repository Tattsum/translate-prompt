import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { analyze, estimate, optimize } from '../api/client'
import { useApp } from '../context/AppContext'
import { Page } from '../components/Layout'

export function InputPage() {
  const { prompt, setPrompt, config, setAnalyzeResult, setOptimizeResult } = useApp()
  const [tokens, setTokens] = useState<number | null>(null)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    const t = setTimeout(() => {
      if (!prompt.trim()) {
        setTokens(null)
        return
      }
      estimate(prompt, config.tokenizer)
        .then((r) => setTokens(r.tokens))
        .catch(() => setTokens(null))
    }, 300)
    return () => clearTimeout(t)
  }, [prompt, config.tokenizer])

  const overLimit = tokens !== null && tokens > config.max_tokens

  async function handleAnalyze() {
    setLoading(true)
    try {
      const result = await analyze(prompt, config)
      setAnalyzeResult(result)
      if (result.status === 'needs_input') {
        navigate('/intake')
      } else {
        const optimized = await optimize(prompt, config)
        setOptimizeResult(optimized)
        navigate('/result')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <Page
      title="プロンプト入力"
      description="最適化したいプロンプトを貼り付けて、分析を開始してください。"
    >
      <div className="card">
        <div className="field">
          <label className="field__label" htmlFor="prompt-input">
            プロンプト
          </label>
          <textarea
            id="prompt-input"
            rows={14}
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="最適化するプロンプトを入力..."
          />
          {tokens !== null && (
            <span className={`token-badge${overLimit ? ' token-badge--warn' : ''}`}>
              推定 {tokens.toLocaleString()} トークン / 上限 {config.max_tokens.toLocaleString()}
            </span>
          )}
        </div>
        <div className="btn-row">
          <button onClick={handleAnalyze} disabled={loading || !prompt.trim()}>
            {loading && <span className="spinner" aria-hidden="true" />}
            {loading ? '分析中...' : '分析して最適化'}
          </button>
          <Link to="/settings" className="btn secondary">
            設定
          </Link>
        </div>
      </div>
    </Page>
  )
}
