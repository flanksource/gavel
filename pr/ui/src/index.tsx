import './index.css';
import { createRoot } from 'react-dom/client';
import { ThemeProvider } from '@flanksource/clicky-ui/hooks';
import { App } from './App';

createRoot(document.getElementById('root')!).render(
  <ThemeProvider defaultTheme="system">
    <App />
  </ThemeProvider>,
);
