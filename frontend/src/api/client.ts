import { createPromiseClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { createClient as createUrqlClient, fetchExchange } from 'urql'

import { TranslatePromptService } from '../gen/translate_prompt/v1/service_connect'
import {
  AnalyzeDocument,
  EstimateDocument,
  InvestigateDocument,
  TargetProfile as GqlTargetProfile,
} from '../gen/graphql/graphql'
import type { AnalyzeResponse, OptimizeConfig, OptimizeResponse, TargetProfile } from './types'

export type { TargetProfile, OptimizeConfig, AnalyzeResponse, OptimizeResponse } from './types'

const apiBase = import.meta.env.VITE_API_BASE_URL ?? ''

const graphqlClient = createUrqlClient({
  url: `${apiBase}/query`,
  exchanges: [fetchExchange],
})

const connectTransport = createConnectTransport({ baseUrl: apiBase || '/' })
const connectClient = createPromiseClient(TranslatePromptService, connectTransport)

function toGraphQLProfile(profile: TargetProfile): GqlTargetProfile {
  const map: Record<TargetProfile, GqlTargetProfile> = {
    claude: GqlTargetProfile.Claude,
    codex: GqlTargetProfile.Codex,
    openai: GqlTargetProfile.Openai,
    devin: GqlTargetProfile.Devin,
    cursor: GqlTargetProfile.Cursor,
  }
  return map[profile]
}

function toGraphQLConfig(config: OptimizeConfig) {
  return {
    targetProfile: toGraphQLProfile(config.target_profile),
    maxTokens: config.max_tokens,
    tokenizer: config.tokenizer,
    deepDive: config.deep_dive,
    workspacePath: config.workspace_path,
  }
}

export async function health() {
  const result = await connectClient.health({})
  return { status: result.status }
}

export async function estimate(text: string, tokenizer: string) {
  const result = await graphqlClient.query(EstimateDocument, { text, tokenizer }).toPromise()
  if (result.error) throw new Error(result.error.message)
  return { tokens: result.data?.estimate.tokens ?? 0 }
}

export async function analyze(prompt: string, config: OptimizeConfig): Promise<AnalyzeResponse> {
  const result = await graphqlClient
    .mutation(AnalyzeDocument, {
      input: { prompt, config: toGraphQLConfig(config) },
    })
    .toPromise()
  if (result.error) throw new Error(result.error.message)
  const data = result.data!.analyze
  return {
    status: data.status === 'NEEDS_INPUT' ? 'needs_input' : 'ready',
    questions: data.questions?.map((q) => ({
      id: q.id,
      text: q.text,
      rule_id: q.ruleId ?? undefined,
    })),
    prompt: data.prompt ?? undefined,
    findings: data.findings?.map((f) => ({
      id: f.id,
      category: f.category,
      severity: f.severity,
      section_id: f.sectionId ?? undefined,
      section_type: f.sectionType ?? undefined,
      rule_id: f.ruleId ?? undefined,
      summary: f.summary,
      source: f.source === 'LLM' ? 'llm' : 'heuristic',
    })),
  }
}

export async function investigate(workspacePath: string, targetProfile: TargetProfile) {
  const result = await graphqlClient
    .mutation(InvestigateDocument, {
      input: {
        workspacePath,
        targetProfile: toGraphQLProfile(targetProfile),
      },
    })
    .toPromise()
  if (result.error) throw new Error(result.error.message)
  const data = result.data!.investigate
  return {
    files: data.files.map((f) => ({
      path: f.path,
      section_type: f.sectionType,
      content_preview: f.contentPreview,
    })),
    suggested_commands: data.suggestedCommands,
  }
}

export async function optimize(
  prompt: string,
  config: OptimizeConfig,
  answers: Record<string, string> = {},
): Promise<OptimizeResponse> {
  const result = await connectClient.optimize({
    prompt,
    config: {
      targetProfile: config.target_profile,
      maxTokens: config.max_tokens,
      tokenizer: config.tokenizer,
      deepDive: config.deep_dive ?? false,
      workspacePath: config.workspace_path ?? '',
    },
    answers,
  })
  return {
    optimized_prompt: result.optimizedPrompt,
    artifacts: {
      cursor_mdc_suggestions: result.artifacts?.cursorMdcSuggestions?.map((m) => ({
        filename: m.filename,
        content: m.content,
      })),
    },
    report: {
      input_tokens: result.report?.inputTokens ?? 0,
      output_tokens: result.report?.outputTokens ?? 0,
      reduction_percent: result.report?.reductionPercent ?? 0,
      target_profile: result.report?.targetProfile ?? '',
      applied_rules:
        result.report?.appliedRules?.map((r) => ({
          id: r.id,
          source_url: r.sourceUrl,
          tokens_delta: r.tokensDelta,
        })) ?? [],
      truncated_sections: result.report?.truncatedSections ?? [],
      stage_results: [],
    },
  }
}
