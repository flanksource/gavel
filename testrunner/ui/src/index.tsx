import { createRoot } from 'react-dom/client';
import styles from './index.css?inline';
import { App } from './App';

// The bundle is inlined into the page by the Go server (and by hosts that embed
// it), so it brings its own stylesheet: clicky-ui's tokens and the utilities
// this package's components use, compiled by Tailwind at build time.
const stylesheet = document.createElement('style');
stylesheet.textContent = styles;
document.head.appendChild(stylesheet);

createRoot(document.getElementById('root')!).render(<App />);
