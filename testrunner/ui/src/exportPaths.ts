import type { Test } from './types';

// The server addresses a node of the exported tree by the slugs of its display
// names, joined from the root (testrunner/ui/routes.go displayTestName,
// slugify and annotateViewPaths). The view's own route paths come from the
// collapsed tree, so a node's export is requested by the path computed here.

const GO_TEST_FRAMEWORK = 'go test';

function insertSpaces(name: string): string {
  let out = '';
  for (let i = 0; i < name.length; i++) {
    const char = name[i]!;
    const prev = i > 0 ? name[i - 1]! : '';
    if (i > 0 && /\p{Lu}/u.test(char) && (/\p{Ll}/u.test(prev) || /\p{Nd}/u.test(prev))) out += ' ';
    out += char;
  }
  return out;
}

function displayTestName(name: string, framework: string | undefined): string {
  if (framework !== GO_TEST_FRAMEWORK || name.endsWith('/')) return name;
  return name
    .split('/')
    .map((part, i) => (i === 0 ? insertSpaces(part.replace(/^Test/, '')) : part.replaceAll('_', ' ')))
    .join(' / ');
}

function slugify(input: string): string {
  return input.toLowerCase().replace(/[^\p{L}\p{N}]+/gu, '-').replace(/^-+|-+$/g, '') || 'node';
}

/** Annotates every node with the path the server's export of that node is requested at. */
export function withExportPaths(tests: Test[], parent: string[] = []): Test[] {
  const slugs = tests.map(test => slugify(displayTestName(test.name, test.framework)));
  const counts = new Map<string, number>();
  slugs.forEach(slug => counts.set(slug, (counts.get(slug) ?? 0) + 1));
  const seen = new Map<string, number>();
  return tests.map((test, i) => {
    const slug = slugs[i]!;
    const ordinal = (seen.get(slug) ?? 0) + 1;
    seen.set(slug, ordinal);
    const path = [...parent, counts.get(slug)! > 1 ? `${slug}~${ordinal}` : slug];
    return {
      ...test,
      export_path: path.join('/'),
      ...(test.children ? { children: withExportPaths(test.children, path) } : {}),
    };
  });
}
