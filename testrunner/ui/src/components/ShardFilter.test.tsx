import { afterEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { ShardFilter, type ResultsShards } from './ShardFilter';

afterEach(cleanup);

const repoBase = '/results/acme/widgets';
const search = '?backTo=%2Fprs%2Facme%2Fwidgets%2F7&shard=11:lint&shard=12:test-pg15&shard=13:test-pg16';

const page = (selected: number[]): ResultsShards => ({
  base: repoBase,
  shards: [
    { artifactId: 11, label: 'lint', selected: selected.includes(11) },
    { artifactId: 12, label: 'test-pg15', selected: selected.includes(12) },
    { artifactId: 13, label: 'test-pg16', selected: selected.includes(13) },
  ],
});

const renderFilter = (selected: number[]) => {
  const basePath = `${repoBase}/${selected.join(',')}`;
  render(<ShardFilter page={page(selected)} location={{ basePath, pathname: `${basePath}/lint`, search }} />);
};

const hrefOf = (name: string) => screen.getByText(name).getAttribute('href');

describe('ShardFilter', () => {
  it('toggles a shard in or out of the merged view, keeping the tab and the shard list', () => {
    renderFilter([11, 12]);

    expect({
      all: hrefOf('All'),
      dropLint: hrefOf('lint'),
      addPg16: hrefOf('test-pg16'),
    }).toEqual({
      all: `${repoBase}/11,12,13/lint${search}`,
      dropLint: `${repoBase}/12/lint${search}`,
      addPg16: `${repoBase}/11,12,13/lint${search}`,
    });
  });

  it('does not offer removing the only selected shard', () => {
    renderFilter([12]);

    expect(hrefOf('test-pg15')).toBeNull();
    expect(hrefOf('lint')).toBe(`${repoBase}/11,12/lint${search}`);
  });
});
