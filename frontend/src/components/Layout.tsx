import type { ReactNode } from 'react'
import { NavLink, Outlet } from 'react-router-dom'

const navItems = [
  { to: '/', label: '入力', end: true },
  { to: '/intake', label: '深堀り' },
  { to: '/result', label: '結果' },
  { to: '/settings', label: '設定' },
] as const

export function AppLayout() {
  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="app-header__inner">
          <NavLink to="/" className="brand" end>
            <span className="brand__icon" aria-hidden="true">
              ✦
            </span>
            <span className="brand__text">
              <span className="brand__title">translate-prompt</span>
              <span className="brand__subtitle">プロンプト最適化</span>
            </span>
          </NavLink>
          <nav className="app-nav" aria-label="メイン">
            {navItems.map(({ to, label, ...rest }) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  `app-nav__link${isActive ? ' app-nav__link--active' : ''}`
                }
                {...rest}
              >
                {label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>
      <Outlet />
    </div>
  )
}

export function Page({ title, description, children }: {
  title: string
  description?: string
  children: ReactNode
}) {
  return (
    <main className="page">
      <div className="page__header">
        <h1 className="page__title">{title}</h1>
        {description && <p className="page__description">{description}</p>}
      </div>
      {children}
    </main>
  )
}
