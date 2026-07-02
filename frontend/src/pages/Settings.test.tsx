import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { AppProvider } from '../context/AppContext'
import { SettingsPage } from './Settings'

describe('SettingsPage', () => {
  it('renders profile selector', () => {
    render(
      <AppProvider>
        <BrowserRouter>
          <SettingsPage />
        </BrowserRouter>
      </AppProvider>,
    )
    expect(screen.getByText('Target Profile')).toBeTruthy()
    expect(screen.getByDisplayValue('codex')).toBeTruthy()
  })
})
