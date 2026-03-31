import { useEffect, useState } from 'react';
import { Save, RotateCcw, CheckCircle2, AlertCircle, Server, Cloud, Shield, Clock } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { AppConfig } from '@/types/models';

const isWails = typeof window !== 'undefined' && 'go' in window;

async function loadConfig(): Promise<AppConfig> {
  if (isWails) {
    const { GetConfig } = await import('../../wailsjs/go/main/App');
    return GetConfig();
  }
  // Dev fallback
  return {
    llm: {
      backend: 'ollama',
      ollama_url: 'http://localhost:11434',
      ollama_model: 'llama3.2',
      cloud_api_key: '',
      cloud_url: 'https://api.anthropic.com',
      cloud_model: 'claude-sonnet-4-6',
      max_tokens: 4096,
      temperature: 0.3,
    },
    security: {
      allow_remediation: false,
      allowed_remediations: ['block_port', 'kill_process'],
      scan_interval_seconds: 300,
    },
  };
}

async function saveConfig(config: AppConfig): Promise<void> {
  if (isWails) {
    const { SaveConfig } = await import('../../wailsjs/go/main/App');
    return SaveConfig(config);
  }
}

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error';

export function Settings() {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle');
  const [error, setError] = useState('');

  useEffect(() => {
    loadConfig().then(setConfig);
  }, []);

  const handleSave = async () => {
    if (!config) return;
    setSaveStatus('saving');
    setError('');
    try {
      await saveConfig(config);
      setSaveStatus('saved');
      setTimeout(() => setSaveStatus('idle'), 2000);
    } catch (err) {
      setSaveStatus('error');
      setError(err instanceof Error ? err.message : 'Failed to save');
    }
  };

  const handleReset = () => {
    loadConfig().then((cfg) => {
      setConfig(cfg);
      setSaveStatus('idle');
      setError('');
    });
  };

  const updateLLM = (patch: Partial<AppConfig['llm']>) => {
    if (!config) return;
    setConfig({ ...config, llm: { ...config.llm, ...patch } });
    setSaveStatus('idle');
  };

  const updateSecurity = (patch: Partial<AppConfig['security']>) => {
    if (!config) return;
    setConfig({ ...config, security: { ...config.security, ...patch } });
    setSaveStatus('idle');
  };

  if (!config) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-muted animate-pulse">Loading configuration...</div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6 max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-text-primary">Settings</h1>
          <p className="text-sm text-text-secondary">Configure LLM backend and security options</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleReset}
            className="flex items-center gap-1.5 text-xs text-text-muted hover:text-text-primary px-3 py-2 rounded-lg border border-border-default hover:border-border-active transition-colors"
          >
            <RotateCcw size={12} />
            Reset
          </button>
          <button
            onClick={handleSave}
            disabled={saveStatus === 'saving'}
            className={cn(
              'flex items-center gap-1.5 text-xs px-3 py-2 rounded-lg border transition-all',
              saveStatus === 'saved'
                ? 'border-neon-green/30 text-neon-green bg-neon-green/5'
                : saveStatus === 'error'
                  ? 'border-severity-critical/30 text-severity-critical bg-severity-critical/5'
                  : 'border-neon-cyan/30 text-neon-cyan bg-neon-cyan/5 hover:bg-neon-cyan/10',
            )}
          >
            {saveStatus === 'saved' ? <CheckCircle2 size={12} /> : saveStatus === 'error' ? <AlertCircle size={12} /> : <Save size={12} />}
            {saveStatus === 'saving' ? 'Saving...' : saveStatus === 'saved' ? 'Saved' : saveStatus === 'error' ? 'Error' : 'Save'}
          </button>
        </div>
      </div>

      {error && (
        <div className="text-xs text-severity-critical bg-severity-critical/5 border border-severity-critical/20 rounded-lg px-3 py-2">
          {error}
        </div>
      )}

      {/* LLM Backend Section */}
      <section className="bg-bg-card border border-border-default rounded-xl p-6 space-y-5">
        <h2 className="text-sm font-semibold text-text-primary flex items-center gap-2">
          <Server size={16} className="text-neon-cyan" />
          LLM Backend
        </h2>

        {/* Backend toggle */}
        <div className="space-y-2">
          <label className="text-xs text-text-secondary">Provider</label>
          <div className="grid grid-cols-2 gap-2">
            <BackendButton
              active={config.llm.backend === 'ollama'}
              onClick={() => updateLLM({ backend: 'ollama' })}
              icon={<Server size={14} />}
              label="Ollama (Local)"
              description="Private, runs on your machine"
            />
            <BackendButton
              active={config.llm.backend === 'cloud'}
              onClick={() => updateLLM({ backend: 'cloud' })}
              icon={<Cloud size={14} />}
              label="Cloud (Pro)"
              description="Anthropic/OpenAI API"
            />
          </div>
        </div>

        {/* Ollama settings */}
        {config.llm.backend === 'ollama' && (
          <div className="space-y-3 pl-1">
            <InputField
              label="Ollama URL"
              value={config.llm.ollama_url}
              onChange={(v) => updateLLM({ ollama_url: v })}
              placeholder="http://localhost:11434"
            />
            <InputField
              label="Model"
              value={config.llm.ollama_model}
              onChange={(v) => updateLLM({ ollama_model: v })}
              placeholder="llama3.2"
            />
          </div>
        )}

        {/* Cloud settings */}
        {config.llm.backend === 'cloud' && (
          <div className="space-y-3 pl-1">
            <InputField
              label="API Key"
              value={config.llm.cloud_api_key ?? ''}
              onChange={(v) => updateLLM({ cloud_api_key: v })}
              placeholder="sk-..."
              type="password"
            />
            <InputField
              label="API Base URL"
              value={config.llm.cloud_url ?? ''}
              onChange={(v) => updateLLM({ cloud_url: v })}
              placeholder="https://api.anthropic.com"
            />
            <InputField
              label="Model"
              value={config.llm.cloud_model ?? ''}
              onChange={(v) => updateLLM({ cloud_model: v })}
              placeholder="claude-sonnet-4-6"
            />
          </div>
        )}

        {/* Shared LLM settings */}
        <div className="grid grid-cols-2 gap-3 pt-2 border-t border-border-default">
          <div className="space-y-1.5">
            <label className="text-xs text-text-secondary">Max Tokens</label>
            <input
              type="number"
              value={config.llm.max_tokens}
              onChange={(e) => updateLLM({ max_tokens: parseInt(e.target.value) || 4096 })}
              className="w-full bg-bg-primary border border-border-default rounded-lg px-3 py-2 text-sm text-text-primary focus:outline-none focus:border-neon-cyan/50"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs text-text-secondary">Temperature</label>
            <input
              type="number"
              step="0.1"
              min="0"
              max="2"
              value={config.llm.temperature}
              onChange={(e) => updateLLM({ temperature: parseFloat(e.target.value) || 0.3 })}
              className="w-full bg-bg-primary border border-border-default rounded-lg px-3 py-2 text-sm text-text-primary focus:outline-none focus:border-neon-cyan/50"
            />
          </div>
        </div>
      </section>

      {/* Security Section */}
      <section className="bg-bg-card border border-border-default rounded-xl p-6 space-y-5">
        <h2 className="text-sm font-semibold text-text-primary flex items-center gap-2">
          <Shield size={16} className="text-neon-cyan" />
          Security
        </h2>

        {/* Remediation toggle */}
        <div className="flex items-center justify-between">
          <div>
            <span className="text-sm text-text-primary">Allow Remediation Actions</span>
            <p className="text-xs text-text-muted mt-0.5">
              Let the agent execute safe commands (block port, kill process)
            </p>
          </div>
          <ToggleSwitch
            enabled={config.security.allow_remediation}
            onChange={(v) => updateSecurity({ allow_remediation: v })}
          />
        </div>

        {/* Allowed remediations */}
        {config.security.allow_remediation && (
          <div className="space-y-2 pl-1">
            <label className="text-xs text-text-secondary">Allowed Actions</label>
            <div className="flex flex-wrap gap-2">
              {['block_port', 'kill_process', 'disable_service'].map((action) => {
                const enabled = config.security.allowed_remediations.includes(action);
                return (
                  <button
                    key={action}
                    onClick={() => {
                      const next = enabled
                        ? config.security.allowed_remediations.filter((a) => a !== action)
                        : [...config.security.allowed_remediations, action];
                      updateSecurity({ allowed_remediations: next });
                    }}
                    className={cn(
                      'text-xs px-3 py-1.5 rounded-lg border transition-colors',
                      enabled
                        ? 'border-neon-cyan/30 text-neon-cyan bg-neon-cyan/5'
                        : 'border-border-default text-text-muted hover:text-text-secondary',
                    )}
                  >
                    {action.replace('_', ' ')}
                  </button>
                );
              })}
            </div>
          </div>
        )}

        {/* Scan interval */}
        <div className="flex items-center justify-between pt-3 border-t border-border-default">
          <div className="flex items-center gap-2">
            <Clock size={14} className="text-text-muted" />
            <div>
              <span className="text-sm text-text-primary">Scan Interval</span>
              <p className="text-xs text-text-muted mt-0.5">How often the background scanner runs</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <input
              type="number"
              min="30"
              max="3600"
              value={config.security.scan_interval_seconds}
              onChange={(e) => updateSecurity({ scan_interval_seconds: parseInt(e.target.value) || 300 })}
              className="w-20 bg-bg-primary border border-border-default rounded-lg px-3 py-1.5 text-sm text-text-primary text-right focus:outline-none focus:border-neon-cyan/50"
            />
            <span className="text-xs text-text-muted">sec</span>
          </div>
        </div>
      </section>
    </div>
  );
}

function BackendButton({ active, onClick, icon, label, description }: {
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  label: string;
  description: string;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex flex-col items-start p-3 rounded-lg border text-left transition-all',
        active
          ? 'border-neon-cyan/40 bg-neon-cyan/5 text-neon-cyan'
          : 'border-border-default bg-bg-secondary/50 text-text-secondary hover:border-border-active hover:text-text-primary',
      )}
    >
      <div className="flex items-center gap-2 mb-1">
        {icon}
        <span className="text-sm font-medium">{label}</span>
      </div>
      <span className="text-xs text-text-muted">{description}</span>
    </button>
  );
}

function InputField({ label, value, onChange, placeholder, type = 'text' }: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder: string;
  type?: string;
}) {
  return (
    <div className="space-y-1.5">
      <label className="text-xs text-text-secondary">{label}</label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full bg-bg-primary border border-border-default rounded-lg px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-neon-cyan/50 focus:ring-1 focus:ring-neon-cyan/20 transition-colors"
      />
    </div>
  );
}

function ToggleSwitch({ enabled, onChange }: { enabled: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      onClick={() => onChange(!enabled)}
      className={cn(
        'relative w-10 h-5 rounded-full transition-colors',
        enabled ? 'bg-neon-cyan/30' : 'bg-bg-primary border border-border-default',
      )}
    >
      <div
        className={cn(
          'absolute top-0.5 w-4 h-4 rounded-full transition-all',
          enabled ? 'left-5 bg-neon-cyan' : 'left-0.5 bg-text-muted',
        )}
      />
    </button>
  );
}
