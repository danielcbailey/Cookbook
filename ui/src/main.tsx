import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { deleteParameterByName, getParameterByName } from './helpers.ts'
import { setToken } from './shared/authHelpers.ts'

const token = getParameterByName('token');
if (token) {
  deleteParameterByName('token');
  setToken(token);
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
