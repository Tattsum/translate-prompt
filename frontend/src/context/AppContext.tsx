import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'
import type { AnalyzeResponse, OptimizeConfig, OptimizeResponse } from '../api/types'

interface AppState {
  prompt: string
  setPrompt: (v: string) => void
  config: OptimizeConfig
  setConfig: (v: OptimizeConfig) => void
  analyzeResult: AnalyzeResponse | null
  setAnalyzeResult: (v: AnalyzeResponse | null) => void
  answers: Record<string, string>
  setAnswers: (v: Record<string, string>) => void
  optimizeResult: OptimizeResponse | null
  setOptimizeResult: (v: OptimizeResponse | null) => void
}

const AppContext = createContext<AppState | null>(null)

const defaultConfig: OptimizeConfig = {
  target_profile: 'codex',
  max_tokens: 8000,
  tokenizer: 'cl100k_base',
  deep_dive: true,
}

export function AppProvider({ children }: { children: ReactNode }) {
  const [prompt, setPrompt] = useState('')
  const [config, setConfig] = useState<OptimizeConfig>(defaultConfig)
  const [analyzeResult, setAnalyzeResult] = useState<AnalyzeResponse | null>(null)
  const [answers, setAnswers] = useState<Record<string, string>>({})
  const [optimizeResult, setOptimizeResult] = useState<OptimizeResponse | null>(null)

  const value = useMemo(
    () => ({
      prompt,
      setPrompt,
      config,
      setConfig,
      analyzeResult,
      setAnalyzeResult,
      answers,
      setAnswers,
      optimizeResult,
      setOptimizeResult,
    }),
    [prompt, config, analyzeResult, answers, optimizeResult],
  )

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>
}

export function useApp() {
  const ctx = useContext(AppContext)
  if (!ctx) throw new Error('useApp must be used within AppProvider')
  return ctx
}
