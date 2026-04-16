export interface ParsedGoroutineFrame {
  functionName: string;
  displayName: string;
  file?: string;
  location?: string;
  line?: number;
  kind: 'frame' | 'created_by';
  runtime: boolean;
}

export interface ParsedGoroutine {
  id: number;
  state: string;
  rawState: string;
  frames: ParsedGoroutineFrame[];
  raw: string;
  userFrameCount: number;
  topFunction?: string;
  searchText: string;
}

const headerRe = /^goroutine\s+(\d+)\s+\[(.+?)\]:$/;
const fileRe = /^\s*(.+?):(\d+)(?:\s+\+0x[0-9a-f]+)?$/i;
const goSrcPrefixRe = /^\/usr\/local\/go[\d.]+\/src\//;
const goWorkspacePrefixRe = /^.*?\/go\/src\//;
const goPkgModPrefixRe = /^.*?\/go\/pkg\/mod\//;

export function parseGoroutineDump(text: string): ParsedGoroutine[] {
  const trimmed = text.trim();
  if (!trimmed) return [];

  const blocks = trimmed.split(/\n\s*\n+/);
  const goroutines: ParsedGoroutine[] = [];

  for (const block of blocks) {
    const lines = block.split('\n').map(line => line.replace(/\r$/, ''));
    const header = lines[0]?.trim();
    const match = headerRe.exec(header || '');
    if (!match) continue;

    const id = Number(match[1]);
    const rawState = match[2];
    const state = normalizeState(rawState);
    const frames: ParsedGoroutineFrame[] = [];

    for (let i = 1; i < lines.length; i++) {
      const line = lines[i];
      const trimmedLine = line.trim();
      if (!trimmedLine) continue;

      const fileMatch = fileRe.exec(trimmedLine);
      if (fileMatch && frames.length > 0) {
        frames[frames.length - 1].file = fileMatch[1];
        frames[frames.length - 1].line = Number(fileMatch[2]);
        continue;
      }

      const kind = trimmedLine.startsWith('created by ') ? 'created_by' : 'frame';
      const functionName = kind === 'created_by'
        ? trimmedLine.slice('created by '.length)
        : trimmedLine;
      frames.push({
        functionName,
        displayName: sanitizeFunctionName(functionName),
        kind,
        runtime: isRuntimeFrame(functionName),
      });
    }

    for (const frame of frames) {
      if (frame.file) {
        frame.file = normalizeFilePath(frame.file);
        frame.location = `${frame.file}${frame.line ? `:${frame.line}` : ''}`;
      }
    }

    goroutines.push({
      id,
      state,
      rawState,
      frames,
      raw: block,
      userFrameCount: frames.filter(frame => !frame.runtime && frame.kind === 'frame').length,
      topFunction: frames.find(frame => frame.kind === 'frame')?.functionName,
      searchText: `${header}\n${frames.map(frame => `${frame.functionName} ${frame.file || ''}`).join('\n')}`.toLowerCase(),
    });
  }

  return goroutines;
}

export function countGoroutinesByState(goroutines: ParsedGoroutine[]): Map<string, number> {
  const counts = new Map<string, number>();
  for (const goroutine of goroutines) {
    counts.set(goroutine.state, (counts.get(goroutine.state) || 0) + 1);
  }
  return counts;
}

function normalizeState(value: string): string {
  return value.split(',')[0].trim().toLowerCase();
}

function isRuntimeFrame(functionName: string): boolean {
  return functionName.startsWith('runtime.') ||
    functionName.startsWith('runtime/') ||
    functionName.startsWith('internal/') ||
    functionName.startsWith('runtime/internal/') ||
    functionName.startsWith('syscall.') ||
    functionName.startsWith('reflect.');
}

function sanitizeFunctionName(functionName: string): string {
  let name = functionName.trim();
  const paren = name.indexOf('(');
  if (paren !== -1) {
    name = name.slice(0, paren);
  }
  name = name.replace(/\.\(\*([^)]+)\)\./g, '.$1.');
  name = stripPackageQualifier(name);
  return name;
}

function stripPackageQualifier(name: string): string {
  return name.replace(/^((?:[^./\s]+\/)+)([^./\s]+)\./, '$2.');
}

function normalizeFilePath(path: string): string {
  let normalized = path
    .replace(goSrcPrefixRe, '')
    .replace(goWorkspacePrefixRe, '')
    .replace(goPkgModPrefixRe, '');
  return normalized;
}
