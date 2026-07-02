import { useNavigate } from 'react-router-dom'
import { useApp } from '../context/AppContext'
import { optimize } from '../api/client'
import { useState } from 'react'
import { Page } from '../components/Layout'

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
    <Page
      title="深堀り (Intake)"
      description="不足している情報を補うための質問に回答してください。"
    >
      {questions.length === 0 ? (
        <div className="card">
          <div className="empty-state">
            <div className="empty-state__icon" aria-hidden="true">
              ?
            </div>
            <p className="empty-state__text">
              質問はありません。入力ページから分析を実行してください。
            </p>
          </div>
        </div>
      ) : (
        <div className="card">
          <div className="question-list">
            {questions.map((q) => (
              <div key={q.id} className="field">
                <label className="field__label" htmlFor={`q-${q.id}`}>
                  <span className="question-item__text">{q.text}</span>
                  {q.rule_id && (
                    <span className="question-item__rule"> {q.rule_id}</span>
                  )}
                </label>
                <input
                  id={`q-${q.id}`}
                  value={answers[q.id] ?? ''}
                  onChange={(e) => setAnswers({ ...answers, [q.id]: e.target.value })}
                />
              </div>
            ))}
          </div>
          <div className="btn-row">
            <button onClick={handleSubmit} disabled={loading}>
              {loading && <span className="spinner" aria-hidden="true" />}
              {loading ? '最適化中...' : '回答を反映して最適化'}
            </button>
          </div>
        </div>
      )}
    </Page>
  )
}
