/* eslint-disable */
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
export type Maybe<T> = T | null;
export type InputMaybe<T> = T | null | undefined;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Map: { input: any; output: any; }
};

export type AnalyzeInput = {
  config: OptimizeConfigInput;
  prompt: Scalars['String']['input'];
};

export type AnalyzeResult = {
  __typename?: 'AnalyzeResult';
  prompt?: Maybe<Scalars['String']['output']>;
  questions?: Maybe<Array<Question>>;
  status: AnalyzeStatus;
};

export enum AnalyzeStatus {
  NeedsInput = 'NEEDS_INPUT',
  Ready = 'READY'
}

export type AppliedRule = {
  __typename?: 'AppliedRule';
  id: Scalars['ID']['output'];
  sourceUrl: Scalars['String']['output'];
  tokensDelta?: Maybe<Scalars['Int']['output']>;
};

export type EstimateResult = {
  __typename?: 'EstimateResult';
  tokens: Scalars['Int']['output'];
};

export type Health = {
  __typename?: 'Health';
  status: Scalars['String']['output'];
};

export type InvestigateInput = {
  targetProfile: TargetProfile;
  workspacePath: Scalars['String']['input'];
};

export type InvestigationFile = {
  __typename?: 'InvestigationFile';
  contentPreview: Scalars['String']['output'];
  path: Scalars['String']['output'];
  sectionType: Scalars['String']['output'];
};

export type InvestigationResult = {
  __typename?: 'InvestigationResult';
  files: Array<InvestigationFile>;
  suggestedCommands: Array<Scalars['String']['output']>;
};

export type MdcSuggestion = {
  __typename?: 'MdcSuggestion';
  content: Scalars['String']['output'];
  filename: Scalars['String']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  analyze: AnalyzeResult;
  investigate: InvestigationResult;
  optimize: OptimizeResult;
};


export type MutationAnalyzeArgs = {
  input: AnalyzeInput;
};


export type MutationInvestigateArgs = {
  input: InvestigateInput;
};


export type MutationOptimizeArgs = {
  input: OptimizeInput;
};

export type OptimizeArtifacts = {
  __typename?: 'OptimizeArtifacts';
  cursorMdcSuggestions?: Maybe<Array<MdcSuggestion>>;
};

export type OptimizeConfigInput = {
  deepDive?: InputMaybe<Scalars['Boolean']['input']>;
  maxTokens: Scalars['Int']['input'];
  targetProfile: TargetProfile;
  tokenizer: Scalars['String']['input'];
  workspacePath?: InputMaybe<Scalars['String']['input']>;
};

export type OptimizeInput = {
  answers?: InputMaybe<Scalars['Map']['input']>;
  config: OptimizeConfigInput;
  prompt: Scalars['String']['input'];
};

export type OptimizeReport = {
  __typename?: 'OptimizeReport';
  appliedRules: Array<AppliedRule>;
  inputTokens: Scalars['Int']['output'];
  outputTokens: Scalars['Int']['output'];
  reductionPercent: Scalars['Float']['output'];
  targetProfile: Scalars['String']['output'];
  truncatedSections: Array<Scalars['String']['output']>;
};

export type OptimizeResult = {
  __typename?: 'OptimizeResult';
  artifacts: OptimizeArtifacts;
  optimizedPrompt: Scalars['String']['output'];
  report: OptimizeReport;
};

export type Query = {
  __typename?: 'Query';
  estimate: EstimateResult;
  health: Health;
};


export type QueryEstimateArgs = {
  text: Scalars['String']['input'];
  tokenizer: Scalars['String']['input'];
};

export type Question = {
  __typename?: 'Question';
  id: Scalars['ID']['output'];
  ruleId?: Maybe<Scalars['String']['output']>;
  text: Scalars['String']['output'];
};

export enum TargetProfile {
  Claude = 'CLAUDE',
  Codex = 'CODEX',
  Cursor = 'CURSOR',
  Devin = 'DEVIN',
  Openai = 'OPENAI'
}

export type AnalyzeMutationVariables = Exact<{
  input: AnalyzeInput;
}>;


export type AnalyzeMutation = { __typename?: 'Mutation', analyze: { __typename?: 'AnalyzeResult', status: AnalyzeStatus, prompt?: string | null, questions?: Array<{ __typename?: 'Question', id: string, text: string, ruleId?: string | null }> | null } };

export type InvestigateMutationVariables = Exact<{
  input: InvestigateInput;
}>;


export type InvestigateMutation = { __typename?: 'Mutation', investigate: { __typename?: 'InvestigationResult', suggestedCommands: Array<string>, files: Array<{ __typename?: 'InvestigationFile', path: string, sectionType: string, contentPreview: string }> } };

export type OptimizeMutationVariables = Exact<{
  input: OptimizeInput;
}>;


export type OptimizeMutation = { __typename?: 'Mutation', optimize: { __typename?: 'OptimizeResult', optimizedPrompt: string, artifacts: { __typename?: 'OptimizeArtifacts', cursorMdcSuggestions?: Array<{ __typename?: 'MdcSuggestion', filename: string, content: string }> | null }, report: { __typename?: 'OptimizeReport', inputTokens: number, outputTokens: number, reductionPercent: number, targetProfile: string, truncatedSections: Array<string>, appliedRules: Array<{ __typename?: 'AppliedRule', id: string, sourceUrl: string, tokensDelta?: number | null }> } } };

export type EstimateQueryVariables = Exact<{
  text: Scalars['String']['input'];
  tokenizer: Scalars['String']['input'];
}>;


export type EstimateQuery = { __typename?: 'Query', estimate: { __typename?: 'EstimateResult', tokens: number } };

export type HealthQueryVariables = Exact<{ [key: string]: never; }>;


export type HealthQuery = { __typename?: 'Query', health: { __typename?: 'Health', status: string } };


export const AnalyzeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Analyze"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"AnalyzeInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"analyze"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}},{"kind":"Field","name":{"kind":"Name","value":"prompt"}},{"kind":"Field","name":{"kind":"Name","value":"questions"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"text"}},{"kind":"Field","name":{"kind":"Name","value":"ruleId"}}]}}]}}]}}]} as unknown as DocumentNode<AnalyzeMutation, AnalyzeMutationVariables>;
export const InvestigateDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Investigate"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"InvestigateInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"investigate"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"files"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"path"}},{"kind":"Field","name":{"kind":"Name","value":"sectionType"}},{"kind":"Field","name":{"kind":"Name","value":"contentPreview"}}]}},{"kind":"Field","name":{"kind":"Name","value":"suggestedCommands"}}]}}]}}]} as unknown as DocumentNode<InvestigateMutation, InvestigateMutationVariables>;
export const OptimizeDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"mutation","name":{"kind":"Name","value":"Optimize"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"input"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"OptimizeInput"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"optimize"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"input"},"value":{"kind":"Variable","name":{"kind":"Name","value":"input"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"optimizedPrompt"}},{"kind":"Field","name":{"kind":"Name","value":"artifacts"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"cursorMdcSuggestions"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"filename"}},{"kind":"Field","name":{"kind":"Name","value":"content"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"report"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"inputTokens"}},{"kind":"Field","name":{"kind":"Name","value":"outputTokens"}},{"kind":"Field","name":{"kind":"Name","value":"reductionPercent"}},{"kind":"Field","name":{"kind":"Name","value":"targetProfile"}},{"kind":"Field","name":{"kind":"Name","value":"appliedRules"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"sourceUrl"}},{"kind":"Field","name":{"kind":"Name","value":"tokensDelta"}}]}},{"kind":"Field","name":{"kind":"Name","value":"truncatedSections"}}]}}]}}]}}]} as unknown as DocumentNode<OptimizeMutation, OptimizeMutationVariables>;
export const EstimateDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Estimate"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"text"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"tokenizer"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"String"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"estimate"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"text"},"value":{"kind":"Variable","name":{"kind":"Name","value":"text"}}},{"kind":"Argument","name":{"kind":"Name","value":"tokenizer"},"value":{"kind":"Variable","name":{"kind":"Name","value":"tokenizer"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"tokens"}}]}}]}}]} as unknown as DocumentNode<EstimateQuery, EstimateQueryVariables>;
export const HealthDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"Health"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"health"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"status"}}]}}]}}]} as unknown as DocumentNode<HealthQuery, HealthQueryVariables>;