import { cn } from '@/lib/utils';

interface ScoreGaugeProps {
  score: number;
  maxScore: number;
  label?: string;
}

function scoreColor(score: number): string {
  if (score < 0) return 'text-text-muted';
  if (score >= 80) return 'text-neon-green';
  if (score >= 60) return 'text-neon-amber';
  if (score >= 40) return 'text-severity-high';
  return 'text-severity-critical';
}

function scoreGlow(score: number): string {
  if (score < 0) return '';
  if (score >= 80) return 'drop-shadow-[0_0_12px_#39ff14]';
  if (score >= 60) return 'drop-shadow-[0_0_12px_#ffab00]';
  if (score >= 40) return 'drop-shadow-[0_0_12px_#ff6b35]';
  return 'drop-shadow-[0_0_12px_#ff3d3d]';
}

function strokeColor(score: number): string {
  if (score < 0) return '#555577';
  if (score >= 80) return '#39ff14';
  if (score >= 60) return '#ffab00';
  if (score >= 40) return '#ff6b35';
  return '#ff3d3d';
}

export function ScoreGauge({ score, maxScore, label = 'Security Score' }: ScoreGaugeProps) {
  const displayScore = score < 0 ? '--' : score;
  const pct = score < 0 ? 0 : (score / maxScore) * 100;

  // SVG circle math
  const radius = 70;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (pct / 100) * circumference;

  return (
    <div className="flex flex-col items-center">
      <div className={cn('relative', scoreGlow(score))}>
        <svg width="180" height="180" viewBox="0 0 180 180">
          {/* Background circle */}
          <circle
            cx="90" cy="90" r={radius}
            fill="none"
            stroke="#2a2a44"
            strokeWidth="8"
          />
          {/* Score arc */}
          <circle
            cx="90" cy="90" r={radius}
            fill="none"
            stroke={strokeColor(score)}
            strokeWidth="8"
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            transform="rotate(-90 90 90)"
            className="transition-all duration-1000 ease-out"
          />
        </svg>
        {/* Score text */}
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className={cn('text-4xl font-bold', scoreColor(score))}>
            {displayScore}
          </span>
          <span className="text-xs text-text-muted mt-1">/ {maxScore}</span>
        </div>
      </div>
      <span className="text-sm text-text-secondary mt-2">{label}</span>
    </div>
  );
}
