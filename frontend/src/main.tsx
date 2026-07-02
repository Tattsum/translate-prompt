import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import './index.css'
import { InputPage } from './pages/Input'
import { SettingsPage } from './pages/Settings'
import { IntakePage } from './pages/Intake'
import { ResultPage } from './pages/Result'
import { AppProvider } from './context/AppContext'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AppProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<InputPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/intake" element={<IntakePage />} />
          <Route path="/result" element={<ResultPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AppProvider>
  </StrictMode>,
)
