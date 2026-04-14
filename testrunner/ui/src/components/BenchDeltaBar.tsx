import type { BenchDelta } from '../types';

interface Props {
  delta: BenchDelta;
  threshold: number;
}

// BenchDeltaBar renders a horizontal bar showing the signed delta percentage.
// Width is proportional to |delta| clamped at 50%. Right = regression (red),
// left = improvement (green). Gray if not statistically significant.
export function BenchDeltaBar({ delta, threshold }: Props) {
  const pct = delta.delta_pct;
  const sig = !!delta.significant;
  const isRegression = sig && pct > threshold;
  const isImprovement = sig && pct < -threshold;

  const clamped = Math.max(-50, Math.min(50, pct));
  const width = Math.abs(clamped) * 2; // 0..100 -> 0..100%
  const color = isRegression ? 'bg-red-500' : isImprovement ? 'bg-green-500' : 'bg-gray-300';
  const textColor = isRegression ? 'text-red-600' : isImprovement ? 'text-green-600' : 'text-gray-500';
  const fontWeight = sig ? 'font-semibold' : '';

  return (
    <div class="flex items-center gap-2 w-full">
      <div class="relative flex-1 h-4 bg-gray-100 rounded overflow-hidden min-w-[80px]">
        <div class="absolute top-0 bottom-0 left-1/2 w-px bg-gray-400" />
        <div
          class={`absolute top-0 bottom-0 ${color}`}
          style={
            pct >= 0
              ? { left: '50%', width: `${width / 2}%` }
              : { right: '50%', width: `${width / 2}%` }
          }
        />
      </div>
      <span class={`text-xs tabular-nums w-16 text-right ${textColor} ${fontWeight}`}>
        {pct >= 0 ? '+' : ''}
        {pct.toFixed(2)}%
      </span>
    </div>
  );
}
