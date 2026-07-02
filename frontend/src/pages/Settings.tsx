import { Nav } from './Input'
import { useApp } from '../context/AppContext'
import type { TargetProfile } from '../api/types'

const profiles: TargetProfile[] = ['claude', 'codex', 'openai', 'devin', 'cursor']

export function SettingsPage() {
  const { config, setConfig } = useApp()

  return (
    <>
      <Nav />
      <main>
        <h1>設定</h1>
        <div className="card">
          <label>
            Target Profile
            <select
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
          </label>
          <label style={{ display: 'block', marginTop: '1rem' }}>
            Max Tokens
            <input
              type="number"
              value={config.max_tokens}
              onChange={(e) => setConfig({ ...config, max_tokens: Number(e.target.value) })}
            />
          </label>
          <label style={{ display: 'block', marginTop: '1rem' }}>
            Tokenizer
            <select
              value={config.tokenizer}
              onChange={(e) => setConfig({ ...config, tokenizer: e.target.value })}
            >
              <option value="cl100k_base">cl100k_base</option>
              <option value="o200k_base">o200k_base</option>
            </select>
          </label>
          <label style={{ display: 'block', marginTop: '1rem' }}>
            <input
              type="checkbox"
              checked={config.deep_dive ?? false}
              onChange={(e) => setConfig({ ...config, deep_dive: e.target.checked })}
            />{' '}
            深堀り (Intake) を有効化
          </label>
          <label style={{ display: 'block', marginTop: '1rem' }}>
            Workspace Path
            <input
              type="text"
              value={config.workspace_path ?? ''}
              onChange={(e) => setConfig({ ...config, workspace_path: e.target.value })}
              placeholder="/path/to/repo"
            />
          </label>
        </div>
      </main>
    </>
  )
}
