import type { SearchConfig } from '../types';

interface Props {
  config: SearchConfig;
  onChange: (partial: Partial<SearchConfig>) => void;
}

export function SearchControls({ config, onChange }: Props) {
  const isMe = !config.any && (config.author === '@me' || config.author === '');
  const isAll = !!config.any;

  return (
    <div class="flex items-center gap-1">
      <button
        class={`text-xs px-2 py-0.5 rounded-full border transition-colors ${
          isMe
            ? 'bg-blue-50 border-blue-300 text-blue-700 font-medium'
            : 'border-gray-200 text-gray-500 hover:bg-gray-50'
        }`}
        onClick={() => onChange({ author: '@me', any: false })}
        title="Show only my PRs"
      >
        <iconify-icon icon="codicon:person" class="mr-0.5 text-[10px]" />
        @me
      </button>
      <button
        class={`text-xs px-2 py-0.5 rounded-full border transition-colors ${
          isAll
            ? 'bg-blue-50 border-blue-300 text-blue-700 font-medium'
            : 'border-gray-200 text-gray-500 hover:bg-gray-50'
        }`}
        onClick={() => onChange({ author: '', any: true })}
        title="Show PRs from all authors"
      >
        <iconify-icon icon="codicon:organization" class="mr-0.5 text-[10px]" />
        All
      </button>
      <span class="text-gray-300 mx-0.5">|</span>
      <button
        class={`text-xs px-2 py-0.5 rounded-full border transition-colors ${
          config.bots
            ? 'bg-gray-100 border-gray-400 text-gray-700 font-medium'
            : 'border-gray-200 text-gray-400 hover:bg-gray-50'
        }`}
        onClick={() => onChange({ bots: !config.bots })}
        title={config.bots ? 'Bots included — click to exclude' : 'Bots excluded — click to include'}
      >
        <iconify-icon icon="codicon:hubot" class="mr-0.5 text-[10px]" />
        Bots
      </button>
    </div>
  );
}
