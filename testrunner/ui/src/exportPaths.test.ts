import { describe, expect, it } from 'vitest';
import { withExportPaths } from './exportPaths';
import type { Test } from './types';

const paths = (tests: Test[]): string[] => tests.flatMap(test => [test.export_path ?? '', ...paths(test.children ?? [])]);

describe('withExportPaths', () => {
  it('names a Go test by the server\'s display label: Test prefix dropped, CamelCase spaced, subtests joined', () => {
    const tree: Test[] = [{
      name: 'pkg/', framework: 'go test',
      children: [
        { name: 'TestParseJSON', framework: 'go test' },
        { name: 'TestRegistry/Detects_Ginkgo', framework: 'go test' },
      ],
    }];

    expect(paths(withExportPaths(tree))).toEqual(['pkg', 'pkg/parse-json', 'pkg/registry-detects-ginkgo']);
  });

  it('slugs the name of any other framework unchanged', () => {
    const tree: Test[] = [{ name: 'parsers/', framework: 'ginkgo', children: [{ name: 'parses valid JSON', framework: 'ginkgo' }, { name: 'TestLooksGo', framework: 'ginkgo' }] }];

    expect(paths(withExportPaths(tree))).toEqual(['parsers', 'parsers/parses-valid-json', 'parsers/testlooksgo']);
  });

  it('numbers siblings that slug alike, in order', () => {
    const tree: Test[] = [{ name: 'a/', children: [{ name: 'same' }, { name: 'Same' }, { name: 'other' }] }];

    expect(paths(withExportPaths(tree))).toEqual(['a', 'a/same~1', 'a/same~2', 'a/other']);
  });

  it('falls back to "node" for a name with no letters or digits', () => {
    expect(paths(withExportPaths([{ name: '---' }]))).toEqual(['node']);
  });
});
