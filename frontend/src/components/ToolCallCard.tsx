import { useState } from 'react';
import { ChevronDown, ChevronRight, CheckCircle2, XCircle, Clock } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { ToolCall } from '@/types/models';

interface ToolCallCardProps {
  toolCall: ToolCall;
}

const toolIcons: Record<string, string> = {
  scan_network: 'Scan Network',
  detect_lateral_movement: 'Lateral Movement',
  analyze_blast_radius: 'Blast Radius',
  inspect_sbom: 'SBOM Inspect',
  remediate: 'Remediate',
};

export function ToolCallCard({ toolCall }: ToolCallCardProps) {
  const [expanded, setExpanded] = useState(false);
  const isSuccess = toolCall.status === 'success';

  return (
    <div className="border border-border-default rounded-lg bg-bg-secondary/50 text-sm my-2">
      {/* Header — always visible */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex items-center gap-2 w-full px-3 py-2 text-left hover:bg-bg-card-hover rounded-lg transition-colors"
      >
        {expanded
          ? <ChevronDown size={14} className="text-text-muted shrink-0" />
          : <ChevronRight size={14} className="text-text-muted shrink-0" />
        }

        {isSuccess
          ? <CheckCircle2 size={14} className="text-neon-green shrink-0" />
          : <XCircle size={14} className="text-severity-critical shrink-0" />
        }

        <span className="text-neon-cyan font-medium">
          {toolIcons[toolCall.name] ?? toolCall.name}
        </span>

        {toolCall.duration_ms != null && (
          <span className="flex items-center gap-1 text-text-muted text-xs ml-auto">
            <Clock size={10} />
            {toolCall.duration_ms}ms
          </span>
        )}
      </button>

      {/* Expanded: args + result */}
      {expanded && (
        <div className="px-3 pb-3 space-y-2 border-t border-border-default pt-2">
          {toolCall.args && (
            <div>
              <span className="text-xs text-text-muted block mb-1">Arguments</span>
              <pre className="text-xs text-text-secondary bg-bg-primary rounded p-2 overflow-x-auto max-h-32 overflow-y-auto">
                {formatJSON(toolCall.args)}
              </pre>
            </div>
          )}
          {toolCall.result && (
            <div>
              <span className="text-xs text-text-muted block mb-1">Result</span>
              <pre className={cn(
                'text-xs rounded p-2 overflow-x-auto max-h-48 overflow-y-auto bg-bg-primary',
                isSuccess ? 'text-text-secondary' : 'text-severity-critical'
              )}>
                {formatJSON(toolCall.result)}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function formatJSON(s: string): string {
  try {
    return JSON.stringify(JSON.parse(s), null, 2);
  } catch {
    return s;
  }
}
