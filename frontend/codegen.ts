import type { CodegenConfig } from '@graphql-codegen/cli'

const config: CodegenConfig = {
  schema: '../backend/graph/schema.graphqls',
  documents: ['src/api/**/*.graphql'],
  generates: {
    './src/gen/graphql/': {
      preset: 'client',
    },
  },
}

export default config
