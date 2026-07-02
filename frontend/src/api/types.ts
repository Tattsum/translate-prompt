export type TargetProfile = 'claude' | 'codex' | 'openai' | 'devin' | 'cursor'

export interface OptimizeConfig {
  target_profile: TargetProfile
  max_tokens: number
  tokenizer: string
  deep_dive?: boolean
  workspace_path?: string
}

export interface Question {
  id: string
  text: string
  rule_id?: string
}

export interface AnalyzeResponse {
  status: 'needs_input' | 'ready'
  questions?: Question[]
  prompt?: string
}

export interface OptimizeReport {
  input_tokens: number
  output_tokens: number
  reduction_percent: number
  target_profile: string
  applied_rules: Array<{ id: string; source_url: string; tokens_delta?: number }>
  truncated_sections: string[]
  stage_results: unknown[]
}

export interface OptimizeResponse {
  optimized_prompt: string
  artifacts: {
    cursor_mdc_suggestions?: Array<{ filename: string; content: string }>
  }
  report: OptimizeReport
}
