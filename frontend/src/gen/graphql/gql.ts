/* eslint-disable */
import * as types from './graphql';
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 * Learn more about it here: https://the-guild.dev/graphql/codegen/plugins/presets/preset-client#reducing-bundle-size
 */
type Documents = {
    "mutation Analyze($input: AnalyzeInput!) {\n  analyze(input: $input) {\n    status\n    prompt\n    questions {\n      id\n      text\n      ruleId\n    }\n    findings {\n      id\n      category\n      severity\n      sectionId\n      sectionType\n      ruleId\n      summary\n      source\n    }\n  }\n}\n\nmutation Investigate($input: InvestigateInput!) {\n  investigate(input: $input) {\n    files {\n      path\n      sectionType\n      contentPreview\n    }\n    suggestedCommands\n  }\n}\n\nmutation Optimize($input: OptimizeInput!) {\n  optimize(input: $input) {\n    optimizedPrompt\n    artifacts {\n      cursorMdcSuggestions {\n        filename\n        content\n      }\n    }\n    report {\n      inputTokens\n      outputTokens\n      reductionPercent\n      targetProfile\n      appliedRules {\n        id\n        sourceUrl\n        tokensDelta\n        method\n        model\n      }\n      truncatedSections\n    }\n  }\n}\n\nquery Estimate($text: String!, $tokenizer: String!) {\n  estimate(text: $text, tokenizer: $tokenizer) {\n    tokens\n  }\n}\n\nquery Health {\n  health {\n    status\n  }\n}": typeof types.AnalyzeDocument,
};
const documents: Documents = {
    "mutation Analyze($input: AnalyzeInput!) {\n  analyze(input: $input) {\n    status\n    prompt\n    questions {\n      id\n      text\n      ruleId\n    }\n    findings {\n      id\n      category\n      severity\n      sectionId\n      sectionType\n      ruleId\n      summary\n      source\n    }\n  }\n}\n\nmutation Investigate($input: InvestigateInput!) {\n  investigate(input: $input) {\n    files {\n      path\n      sectionType\n      contentPreview\n    }\n    suggestedCommands\n  }\n}\n\nmutation Optimize($input: OptimizeInput!) {\n  optimize(input: $input) {\n    optimizedPrompt\n    artifacts {\n      cursorMdcSuggestions {\n        filename\n        content\n      }\n    }\n    report {\n      inputTokens\n      outputTokens\n      reductionPercent\n      targetProfile\n      appliedRules {\n        id\n        sourceUrl\n        tokensDelta\n        method\n        model\n      }\n      truncatedSections\n    }\n  }\n}\n\nquery Estimate($text: String!, $tokenizer: String!) {\n  estimate(text: $text, tokenizer: $tokenizer) {\n    tokens\n  }\n}\n\nquery Health {\n  health {\n    status\n  }\n}": types.AnalyzeDocument,
};

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = graphql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function graphql(source: string): unknown;

/**
 * The graphql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function graphql(source: "mutation Analyze($input: AnalyzeInput!) {\n  analyze(input: $input) {\n    status\n    prompt\n    questions {\n      id\n      text\n      ruleId\n    }\n    findings {\n      id\n      category\n      severity\n      sectionId\n      sectionType\n      ruleId\n      summary\n      source\n    }\n  }\n}\n\nmutation Investigate($input: InvestigateInput!) {\n  investigate(input: $input) {\n    files {\n      path\n      sectionType\n      contentPreview\n    }\n    suggestedCommands\n  }\n}\n\nmutation Optimize($input: OptimizeInput!) {\n  optimize(input: $input) {\n    optimizedPrompt\n    artifacts {\n      cursorMdcSuggestions {\n        filename\n        content\n      }\n    }\n    report {\n      inputTokens\n      outputTokens\n      reductionPercent\n      targetProfile\n      appliedRules {\n        id\n        sourceUrl\n        tokensDelta\n        method\n        model\n      }\n      truncatedSections\n    }\n  }\n}\n\nquery Estimate($text: String!, $tokenizer: String!) {\n  estimate(text: $text, tokenizer: $tokenizer) {\n    tokens\n  }\n}\n\nquery Health {\n  health {\n    status\n  }\n}"): (typeof documents)["mutation Analyze($input: AnalyzeInput!) {\n  analyze(input: $input) {\n    status\n    prompt\n    questions {\n      id\n      text\n      ruleId\n    }\n    findings {\n      id\n      category\n      severity\n      sectionId\n      sectionType\n      ruleId\n      summary\n      source\n    }\n  }\n}\n\nmutation Investigate($input: InvestigateInput!) {\n  investigate(input: $input) {\n    files {\n      path\n      sectionType\n      contentPreview\n    }\n    suggestedCommands\n  }\n}\n\nmutation Optimize($input: OptimizeInput!) {\n  optimize(input: $input) {\n    optimizedPrompt\n    artifacts {\n      cursorMdcSuggestions {\n        filename\n        content\n      }\n    }\n    report {\n      inputTokens\n      outputTokens\n      reductionPercent\n      targetProfile\n      appliedRules {\n        id\n        sourceUrl\n        tokensDelta\n        method\n        model\n      }\n      truncatedSections\n    }\n  }\n}\n\nquery Estimate($text: String!, $tokenizer: String!) {\n  estimate(text: $text, tokenizer: $tokenizer) {\n    tokens\n  }\n}\n\nquery Health {\n  health {\n    status\n  }\n}"];

export function graphql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;