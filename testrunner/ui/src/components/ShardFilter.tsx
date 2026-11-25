export interface ResultsShard {
  artifactId: number;
  label: string;
  selected: boolean;
}

/** The shards the PR dashboard handed the page, as window.__gavelShards. */
export interface ResultsShards {
  /** /results/{repo}: a selection's artifact IDs are appended to it. */
  base: string;
  shards: ResultsShard[];
}

export interface ShardLocation {
  /** The page's base path: {base}/{selected artifact IDs}. */
  basePath: string;
  pathname: string;
  search: string;
}

/** The URL showing artifactIds merged, keeping the current tab and query. */
export function shardSelectionHref(page: ResultsShards, artifactIds: number[], location: ShardLocation): string {
  const rest = location.pathname.startsWith(location.basePath) ? location.pathname.slice(location.basePath.length) : '';
  const ids = page.shards.map(s => s.artifactId).filter(id => artifactIds.includes(id));
  return `${page.base}/${ids.join(',')}${rest}${location.search}`;
}

/** Toggling a shard adds or removes it; the last selected shard cannot be removed. */
export function toggledShards(page: ResultsShards, artifactId: number): number[] | null {
  const selected = page.shards.filter(s => s.selected).map(s => s.artifactId);
  if (!selected.includes(artifactId)) return [...selected, artifactId];
  const rest = selected.filter(id => id !== artifactId);
  return rest.length > 0 ? rest : null;
}

const chip = 'inline-flex items-center rounded-full border px-2 py-0.5 text-xs no-underline';
const on = `${chip} border-primary bg-primary/10 text-primary`;
const off = `${chip} border-border text-muted-foreground hover:bg-muted`;

export function ShardFilter({ page, location }: { page: ResultsShards; location: ShardLocation }) {
  if (page.shards.length < 2) return null;
  const all = page.shards.map(s => s.artifactId);
  const allSelected = page.shards.every(s => s.selected);
  return (
    <nav aria-label="Shards" className="flex min-w-0 flex-wrap items-center gap-1">
      <span className="text-xs text-muted-foreground">Shards</span>
      <a href={shardSelectionHref(page, all, location)} className={allSelected ? on : off} aria-current={allSelected ? 'true' : undefined}>
        All
      </a>
      {page.shards.map(shard => {
        const next = toggledShards(page, shard.artifactId);
        const cls = shard.selected ? on : off;
        return next ? (
          <a key={shard.artifactId} href={shardSelectionHref(page, next, location)} className={cls} aria-current={shard.selected ? 'true' : undefined}>
            {shard.label}
          </a>
        ) : (
          <span key={shard.artifactId} className={cls} aria-current={shard.selected ? 'true' : undefined} title="At least one shard stays selected">
            {shard.label}
          </span>
        );
      })}
    </nav>
  );
}
