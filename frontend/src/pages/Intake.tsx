import { useNavigate } from 'react-router-dom'
import { Nav } from './Input'
import { useApp } from '../context/AppContext'
import { optimize } from '../api/client'
import { useState } from 'react'

export function IntakePage() {
  const { prompt, config, analyzeResult, answers, setAnswers, setOptimizeResult } = useApp()
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const questions = analyzeResult?.questions ?? []

  async function handleSubmit() {
    setLoading(true)
    try {
      const result = await optimize(prompt, config, answers)
      setOptimizeResult(result)
      navigate('/result')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Nav />
      <main>
        <h1>深堀り (Intake)</h1>
        {questions.length === 0 ? (
          <div className="card">
            <p>質問はありません。入力ページから分析を実行してください。</p>
          </div>
        ) : (
          <div className="card">
            {questions.map((q) => (
              <label key={q.id} style={{ display: 'block', marginBottom: '1rem' }}>
                <strong>{q.text}</strong>
                {q.rule_id && (
                  <span style={{ color: '#6b7280', fontSize: '0.85rem' }}> ({q.rule_id})</span>
                )}
                <input
                  style={{ marginTop: '0.25rem' }}
                  value={answers[q.id] ?? ''}
                  onChange={(e) => setAnswers({ ...answers, [q.id]: e.target.value })}
                />
              </label>
            ))}
            <button onClick={handleSubmit} disabled={loading}>
              {loading ? '最適化中...' : '回答を反映して最適化'}
            </button>
          </div>
        )}
      </main>
    </>
  )
}
