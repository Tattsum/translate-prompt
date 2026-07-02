import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { analyze, estimate, optimize } from '../api/client'
import { useApp } from '../context/AppContext'

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
    <>
      <Nav />
      <main>
        <h1>プロンプト入力</h1>
        <div className="card">
          <textarea
            rows={14}
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="最適化するプロンプトを入力..."
          />
          {tokens !== null && (
            <p style={{ marginTop: '0.5rem', color: '#6b7280' }}>
              推定トークン: {tokens} / 上限 {config.max_tokens}
            </p>
          )}
          <div style={{ marginTop: '1rem', display: 'flex', gap: '0.5rem' }}>
            <button onClick={handleAnalyze} disabled={loading || !prompt.trim()}>
              {loading ? '分析中...' : '分析して最適化'}
            </button>
            <Link to="/settings">
              <button type="button" className="secondary">
                設定
              </button>
            </Link>
          </div>
        </div>
      </main>
    </>
  )
}

function Nav() {
  return (
    <nav>
      <Link to="/">入力</Link>
      <Link to="/settings">設定</Link>
      <Link to="/intake">深堀り</Link>
      <Link to="/result">結果</Link>
    </nav>
  )
}

export { Nav }
