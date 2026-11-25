import { cn } from "@/lib/utils";

// Horizontal timeline of nodes with a staggered pulse — suggests commits moving through review.
// Reduced-motion: nodes render at full opacity, no pulse. See globals.css @keyframes pulse-node.
interface PulseNodesProps {
  count?: number;
  labels?: string[];
  className?: string;
}

export default function PulseNodes({ count = 5, labels, className }: PulseNodesProps) {
  const nodes = Array.from({ length: count }, (_, i) => i);

  return (
    <div className={cn("flex h-full w-full items-center justify-center px-4", className)}>
      <svg
        viewBox={`0 0 ${count * 40} 40`}
        preserveAspectRatio="xMidYMid meet"
        className="h-full w-full max-w-sm"
        aria-hidden
      >
        <line
          x1={20}
          y1={20}
          x2={count * 40 - 20}
          y2={20}
          className="stroke-border"
          strokeWidth="1.5"
          strokeDasharray="3 4"
        />
        {nodes.map((i) => (
          <g key={i}>
            <circle
              cx={20 + i * 40}
              cy={20}
              r={4}
              className="fill-brand-sky motion-safe:[animation:pulse-node_1.6s_ease-in-out_infinite]"
              style={{ animationDelay: `${i * 0.2}s` }}
            />
            {labels?.[i] && (
              <text
                x={20 + i * 40}
                y={36}
                textAnchor="middle"
                className="fill-muted-foreground text-[0.5rem] font-mono"
              >
                {labels[i]}
              </text>
            )}
          </g>
        ))}
      </svg>
    </div>
  );
}
