// Minimal ANSI SGR → HTML converter.
// Handles 3/4-bit colors (30-37, 40-47, 90-97, 100-107), bright variants,
// bold/dim/italic/underline attributes, and 256-color / 24-bit sequences
// (38;5;N / 48;5;N / 38;2;r;g;b). Unknown codes reset to default.
// Output is safe to drop into dangerouslySetInnerHTML — text is HTML-escaped.

const ESC = /\x1b\[([\d;]*)m/g;

const FG: Record<number, string> = {
  30: '#000000', 31: '#cd3131', 32: '#0dbc79', 33: '#e5e510',
  34: '#2472c8', 35: '#bc3fbc', 36: '#11a8cd', 37: '#e5e5e5',
  90: '#666666', 91: '#f14c4c', 92: '#23d18b', 93: '#f5f543',
  94: '#3b8eea', 95: '#d670d6', 96: '#29b8db', 97: '#ffffff',
};

const BG: Record<number, string> = {
  40: '#000000', 41: '#cd3131', 42: '#0dbc79', 43: '#e5e510',
  44: '#2472c8', 45: '#bc3fbc', 46: '#11a8cd', 47: '#e5e5e5',
  100: '#666666', 101: '#f14c4c', 102: '#23d18b', 103: '#f5f543',
  104: '#3b8eea', 105: '#d670d6', 106: '#29b8db', 107: '#ffffff',
};

// 256-color palette: first 16 match FG table; 16-231 are 6x6x6 cube;
// 232-255 are grayscale ramp.
function palette256(n: number): string {
  if (n < 8) return FG[30 + n] || '#000';
  if (n < 16) return FG[90 + (n - 8)] || '#fff';
  if (n < 232) {
    const i = n - 16;
    const r = Math.floor(i / 36);
    const g = Math.floor((i % 36) / 6);
    const b = i % 6;
    const conv = (c: number) => (c === 0 ? 0 : 55 + c * 40);
    return `rgb(${conv(r)},${conv(g)},${conv(b)})`;
  }
  const v = 8 + (n - 232) * 10;
  return `rgb(${v},${v},${v})`;
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

interface State {
  fg?: string;
  bg?: string;
  bold?: boolean;
  dim?: boolean;
  italic?: boolean;
  underline?: boolean;
}

function applyCodes(state: State, codes: number[]): State {
  const next = { ...state };
  for (let i = 0; i < codes.length; i++) {
    const c = codes[i];
    if (c === 0 || isNaN(c)) { return {}; }
    else if (c === 1) next.bold = true;
    else if (c === 2) next.dim = true;
    else if (c === 3) next.italic = true;
    else if (c === 4) next.underline = true;
    else if (c === 22) { next.bold = false; next.dim = false; }
    else if (c === 23) next.italic = false;
    else if (c === 24) next.underline = false;
    else if (c === 39) next.fg = undefined;
    else if (c === 49) next.bg = undefined;
    else if (c === 38 && codes[i + 1] === 5 && codes[i + 2] !== undefined) {
      next.fg = palette256(codes[i + 2]);
      i += 2;
    } else if (c === 48 && codes[i + 1] === 5 && codes[i + 2] !== undefined) {
      next.bg = palette256(codes[i + 2]);
      i += 2;
    } else if (c === 38 && codes[i + 1] === 2) {
      next.fg = `rgb(${codes[i + 2] || 0},${codes[i + 3] || 0},${codes[i + 4] || 0})`;
      i += 4;
    } else if (c === 48 && codes[i + 1] === 2) {
      next.bg = `rgb(${codes[i + 2] || 0},${codes[i + 3] || 0},${codes[i + 4] || 0})`;
      i += 4;
    } else if (FG[c]) next.fg = FG[c];
    else if (BG[c]) next.bg = BG[c];
  }
  return next;
}

function openSpan(state: State): string {
  const styles: string[] = [];
  if (state.fg) styles.push(`color:${state.fg}`);
  if (state.bg) styles.push(`background-color:${state.bg}`);
  if (state.bold) styles.push('font-weight:600');
  if (state.dim) styles.push('opacity:0.7');
  if (state.italic) styles.push('font-style:italic');
  if (state.underline) styles.push('text-decoration:underline');
  if (styles.length === 0) return '';
  return `<span style="${styles.join(';')}">`;
}

function isActive(state: State): boolean {
  return !!(state.fg || state.bg || state.bold || state.dim || state.italic || state.underline);
}

// ansiToHtml returns an HTML string with the ANSI coloring preserved.
// The caller is responsible for wrapping in a monospace container.
export function ansiToHtml(input: string): string {
  if (!input) return '';
  let out = '';
  let last = 0;
  let state: State = {};
  let open = false;

  ESC.lastIndex = 0;
  let m: RegExpExecArray | null;
  while ((m = ESC.exec(input)) !== null) {
    if (m.index > last) {
      out += escapeHtml(input.slice(last, m.index));
    }
    const codes = m[1] === '' ? [0] : m[1].split(';').map(n => parseInt(n, 10));
    const nextState = applyCodes(state, codes);
    if (open) { out += '</span>'; open = false; }
    const span = openSpan(nextState);
    if (span) { out += span; open = true; }
    state = nextState;
    last = m.index + m[0].length;
  }
  if (last < input.length) {
    out += escapeHtml(input.slice(last));
  }
  if (open) out += '</span>';
  return out;
}

export function hasAnsi(s: string): boolean {
  return /\x1b\[[\d;]*m/.test(s);
}

export function stripAnsi(s: string): string {
  return s.replace(/\x1b\[[\d;]*m/g, '');
}
