import { useEffect, useState } from 'react';
import { Cpu, HardDrive, AlertTriangle, ShieldCheck, ShieldAlert, Info, Wifi } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ScoreGauge } from './ScoreGauge';
import type { SystemStatus, Finding, Severity } from '@/types/models';

// Stub for when running outside Wails (dev mode)
const isWails = typeof window !== 'undefined' && 'go' in window;

async function getSystemStatus(): Promise<SystemStatus> {
  if (isWails) {
    const { GetSystemStatus } = await import('../../wailsjs/go/main/App');
    return GetSystemStatus();
  }
  // Dev fallback
  return {
    cpu_percent: 23.5,
    memory_used_gb: 8.2,
    memory_total_gb: 16.0,
    network_in_bytes: 0,
    network_out_bytes: 0,
    security_score: { score: 72, max_score: 100, findings: 5, critical: 0, high: 1, medium: 2, low: 2, last_updated: new Date().toISOString() },
    agent_ready: true,
    llm_backend: 'ollama',
  };
}

async function getFindings(): Promise<Finding[]> {
  if (isWails) {
    const { GetFindings } = await import('../../wailsjs/go/main/App');
    return GetFindings() ?? [];
  }
  return [
    { id: 'net-port-22', title: 'SSH listening on port 22', description: 'sshd is accepting connections.', severity: 'medium', category: 'network', source: 'scan_network', timestamp: new Date().toISOString() },
    { id: 'net-port-5432', title: 'PostgreSQL on port 5432', description: 'postgres listening on 0.0.0.0', severity: 'medium', category: 'network', source: 'scan_network', timestamp: new Date().toISOString() },
    { id: 'br-exposed', title: '8 ports exposed to all interfaces', description: 'Multiple services bound to 0.0.0.0', severity: 'high', category: 'blast_radius', source: 'analyze_blast_radius', timestamp: new Date().toISOString() },
  ];
}

const severityIcon: Record<Severity, React.ReactNode> = {
  critical: <ShieldAlert size={14} className="text-severity-critical" />,
  high: <AlertTriangle size={14} className="text-severity-high" />,
  medium: <AlertTriangle size={14} className="text-severity-medium" />,
  low: <ShieldCheck size={14} className="text-severity-low" />,
  info: <Info size={14} className="text-severity-info" />,
};

const severityBorder: Record<Severity, string> = {
  critical: 'border-l-severity-critical',
  high: 'border-l-severity-high',
  medium: 'border-l-severity-medium',
  low: 'border-l-severity-low',
  info: 'border-l-severity-info',
};

export function Dashboard() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [findings, setFindings] = useState<Finding[]>([]);

  useEffect(() => {
    const poll = () => {
      getSystemStatus().then(setStatus);
      getFindings().then(setFindings);
    };
    poll();
    const interval = setInterval(poll, 5000);
    return () => clearInterval(interval);
  }, []);

  if (!status) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-muted animate-pulse">Initializing scanner...</div>
      </div>
    );
  }

  const memPct = status.memory_total_gb > 0
    ? ((status.memory_used_gb / status.memory_total_gb) * 100).toFixed(0)
    : '0';

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-text-primary glow-text-cyan">NeuroPanopticon</h1>
          <p className="text-sm text-text-secondary">Autonomous Security Auditor</p>
        </div>
        <div className="flex items-center gap-3">
          <span className={cn(
            'text-xs px-2 py-1 rounded-full border',
            status.agent_ready
              ? 'border-neon-green/30 text-neon-green bg-neon-green/5'
              : 'border-severity-critical/30 text-severity-critical bg-severity-critical/5'
          )}>
            {status.agent_ready ? 'Agent Online' : 'Agent Offline'}
          </span>
          <span className="text-xs text-text-muted">
            LLM: {status.llm_backend}
          </span>
        </div>
      </div>

      {/* Score + Metrics Row */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Security Score */}
        <div className="bg-bg-card border border-border-default rounded-xl p-6 flex items-center justify-center">
          <ScoreGauge
            score={status.security_score.score}
            maxScore={status.security_score.max_score}
          />
        </div>

        {/* Finding Summary */}
        <div className="bg-bg-card border border-border-default rounded-xl p-6">
          <h2 className="text-sm font-semibold text-text-secondary mb-4">Finding Summary</h2>
          <div className="space-y-3">
            {([
              ['Critical', status.security_score.critical, 'bg-severity-critical'],
              ['High', status.security_score.high, 'bg-severity-high'],
              ['Medium', status.security_score.medium, 'bg-severity-medium'],
              ['Low', status.security_score.low, 'bg-severity-low'],
            ] as const).map(([label, count, color]) => (
              <div key={label} className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className={cn('w-2 h-2 rounded-full', color)} />
                  <span className="text-sm text-text-primary">{label}</span>
                </div>
                <span className="text-sm font-mono text-text-secondary">{count}</span>
              </div>
            ))}
            <div className="pt-2 mt-2 border-t border-border-default flex items-center justify-between">
              <span className="text-sm text-text-secondary">Total</span>
              <span className="text-sm font-mono font-bold text-text-primary">{status.security_score.findings}</span>
            </div>
          </div>
        </div>

        {/* System Metrics */}
        <div className="bg-bg-card border border-border-default rounded-xl p-6">
          <h2 className="text-sm font-semibold text-text-secondary mb-4">System Metrics</h2>
          <div className="space-y-4">
            <MetricBar
              icon={<Cpu size={14} />}
              label="CPU"
              value={`${status.cpu_percent.toFixed(1)}%`}
              pct={status.cpu_percent}
            />
            <MetricBar
              icon={<HardDrive size={14} />}
              label="Memory"
              value={`${status.memory_used_gb.toFixed(1)} / ${status.memory_total_gb.toFixed(1)} GB`}
              pct={parseFloat(memPct)}
            />
            <div className="flex items-center gap-2 pt-2 border-t border-border-default">
              <Wifi size={14} className="text-text-muted" />
              <span className="text-xs text-text-muted">
                Scanner interval: {status.security_score.last_updated
                  ? new Date(status.security_score.last_updated).toLocaleTimeString()
                  : '--'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Findings List */}
      <div className="bg-bg-card border border-border-default rounded-xl p-6">
        <h2 className="text-sm font-semibold text-text-secondary mb-4">
          Recent Findings ({findings.length})
        </h2>
        {findings.length === 0 ? (
          <p className="text-sm text-text-muted py-4 text-center">No findings detected. System looks clean.</p>
        ) : (
          <div className="space-y-2">
            {findings.map((f) => (
              <div
                key={f.id}
                className={cn(
                  'border-l-2 pl-3 py-2 bg-bg-secondary/50 rounded-r-lg',
                  severityBorder[f.severity]
                )}
              >
                <div className="flex items-center gap-2">
                  {severityIcon[f.severity]}
                  <span className="text-sm font-medium text-text-primary">{f.title}</span>
                  <span className="text-xs text-text-muted ml-auto">{f.source}</span>
                </div>
                <p className="text-xs text-text-secondary mt-1 pl-5">{f.description}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function MetricBar({ icon, label, value, pct }: { icon: React.ReactNode; label: string; value: string; pct: number }) {
  const barColor = pct > 90 ? 'bg-severity-critical' : pct > 70 ? 'bg-severity-medium' : 'bg-neon-cyan';
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-text-secondary">
          {icon}
          <span className="text-xs">{label}</span>
        </div>
        <span className="text-xs font-mono text-text-primary">{value}</span>
      </div>
      <div className="h-1.5 bg-bg-primary rounded-full overflow-hidden">
        <div
          className={cn('h-full rounded-full transition-all duration-500', barColor)}
          style={{ width: `${Math.min(pct, 100)}%` }}
        />
      </div>
    </div>
  );
}
