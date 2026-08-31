import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { deleteParameterByName, getParameterByName } from './helpers.ts'

const token = getParameterByName('token');
if (token) {
  deleteParameterByName('token');
  document.cookie = 'session='+token+'; path=/;'
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
