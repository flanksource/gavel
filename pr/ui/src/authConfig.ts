// Loads how the dashboard signs in from /api/auth/config before the app renders.
// It is fetched rather than read from window.__GAVEL__, which dev mode doesn't set.
import type { AuthMode } from './query';

export interface AuthConfig {
  mode: AuthMode;
  publishableKey?: string;
  issuer?: string;
  clientId?: string;
}

export async function loadAuthConfig(): Promise<AuthConfig> {
  const response = await fetch('/api/auth/config');
  if (!response.ok) throw new Error(`load the sign-in settings: HTTP ${response.status}`);
  const config = (await response.json()) as AuthConfig;
  if (config.mode !== 'none' && config.mode !== 'clerk') {
    throw new Error(`load the sign-in settings: unknown sign-in mode "${String(config.mode)}"`);
  }
  return config;
}
