import { describe, expect, it } from 'vitest';
import { chatLayerEnabled } from './ChatLayer';

describe('chatLayerEnabled', () => {
  // The native menubar webview and the focused new-todo window (opened standalone
  // or embedded in the React Grab dialog) are single-purpose surfaces: an
  // assistant FAB would cover their controls.
  it.each(['/menubar', '/todos/new', '/todos/new/'])('hides the chat FAB on %s', pathname => {
    expect(chatLayerEnabled(pathname)).toBe(false);
  });

  it.each(['/', '/todos', '/todos/abc123', '/processes'])('shows the chat FAB on %s', pathname => {
    expect(chatLayerEnabled(pathname)).toBe(true);
  });
});
