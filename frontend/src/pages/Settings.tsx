import { useApp } from '../context/AppContext'
import type { TargetProfile } from '../api/types'
import { Page } from '../components/Layout'

const profiles: TargetProfile[] = ['claude', 'codex', 'openai', 'devin', 'cursor']

const enableWorkspacePath = import.meta.env.VITE_ENABLE_WORKSPACE_PATH === 'true'

export function SettingsPage() {
  const { config, setConfig } = useApp()

  return (
    <Page title="設定" description="最適化のターゲットとトークン制限を調整します。">
      <div className="card">
        <div className="field">
          <label className="field__label" htmlFor="target-profile">
            Target Profile
          </label>
          <select
            id="target-profile"
            value={config.target_profile}
            onChange={(e) =>
              setConfig({ ...config, target_profile: e.target.value as TargetProfile })
            }
          >
            {profiles.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
          <span className="field__hint">最適化先の AI エージェント / モデル向けプロファイル</span>
        </div>

        <div className="field">
          <label className="field__label" htmlFor="max-tokens">
            Max Tokens
          </label>
          <input
            id="max-tokens"
            type="number"
            value={config.max_tokens}
            onChange={(e) => setConfig({ ...config, max_tokens: Number(e.target.value) })}
          />
        </div>

        <div className="field">
          <label className="field__label" htmlFor="tokenizer">
            Tokenizer
          </label>
          <select
            id="tokenizer"
            value={config.tokenizer}
            onChange={(e) => setConfig({ ...config, tokenizer: e.target.value })}
          >
            <option value="cl100k_base">cl100k_base</option>
            <option value="o200k_base">o200k_base</option>
          </select>
        </div>

        <label className="field field--checkbox">
          <input
            type="checkbox"
            checked={config.deep_dive ?? false}
            onChange={(e) => setConfig({ ...config, deep_dive: e.target.checked })}
          />
          <span className="field__label" style={{ margin: 0 }}>
            深堀り (Intake) を有効化
          </span>
        </label>

        {enableWorkspacePath && (
          <div className="field">
            <label className="field__label" htmlFor="workspace-path">
              Workspace Path
            </label>
            <input
              id="workspace-path"
              type="text"
              value={config.workspace_path ?? ''}
              onChange={(e) => setConfig({ ...config, workspace_path: e.target.value })}
              placeholder="/path/to/repo"
            />
          </div>
        )}
      </div>
    </Page>
  )
}
