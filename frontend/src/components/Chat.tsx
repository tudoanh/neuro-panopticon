import { useState, useRef, useEffect } from 'react';
import { Send, RotateCcw, Bot, User, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ToolCallCard } from './ToolCallCard';
import type { ToolCall } from '@/types/models';

const isWails = typeof window !== 'undefined' && 'go' in window;

async function sendMessage(message: string): Promise<{ content: string; tool_calls?: ToolCall[] }> {
  if (isWails) {
    const { SendMessage } = await import('../../wailsjs/go/main/App');
    return SendMessage(message);
  }
  // Dev fallback — simulate a response
  await new Promise((r) => setTimeout(r, 1000));
  return {
    content: `**Analysis complete.** I scanned the system and found 3 listening ports. Port 22 (SSH) is open on all interfaces, which could be a lateral movement vector if compromised.\n\nRecommendation: Restrict SSH to specific IPs using \`ufw allow from <IP> to any port 22\`.`,
    tool_calls: [
      { id: '1', name: 'scan_network', args: '{"include_established": false}', result: '{"listening_ports": [{"port": 22, "process": "sshd"}]}', status: 'success', duration_ms: 142 },
    ],
  };
}

async function resetChat(): Promise<void> {
  if (isWails) {
    const { ResetChat } = await import('../../wailsjs/go/main/App');
    return ResetChat();
  }
}

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  toolCalls?: ToolCall[];
  timestamp: Date;
}

export function Chat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // Auto-scroll on new messages
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, loading]);

  const handleSend = async () => {
    const text = input.trim();
    if (!text || loading) return;

    const userMsg: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content: text,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, userMsg]);
    setInput('');
    setLoading(true);

    try {
      const resp = await sendMessage(text);
      const assistantMsg: Message = {
        id: crypto.randomUUID(),
        role: 'assistant',
        content: resp.content,
        toolCalls: resp.tool_calls,
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, assistantMsg]);
    } catch (err) {
      const errorMsg: Message = {
        id: crypto.randomUUID(),
        role: 'assistant',
        content: `Error: ${err instanceof Error ? err.message : 'Unknown error'}`,
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, errorMsg]);
    } finally {
      setLoading(false);
      inputRef.current?.focus();
    }
  };

  const handleReset = async () => {
    await resetChat();
    setMessages([]);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-3 border-b border-border-default bg-bg-secondary/50">
        <div>
          <h1 className="text-sm font-bold text-text-primary">Security Chat</h1>
          <p className="text-xs text-text-muted">Ask NeuroPanopticon to analyze your system</p>
        </div>
        <button
          onClick={handleReset}
          className="flex items-center gap-1.5 text-xs text-text-muted hover:text-text-primary px-2 py-1 rounded border border-border-default hover:border-border-active transition-colors"
          title="Reset conversation"
        >
          <RotateCcw size={12} />
          Reset
        </button>
      </div>

      {/* Messages area */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
        {messages.length === 0 && !loading && (
          <EmptyState />
        )}

        {messages.map((msg) => (
          <MessageBubble key={msg.id} message={msg} />
        ))}

        {loading && (
          <div className="flex items-center gap-2 text-text-muted text-sm py-2">
            <Loader2 size={14} className="animate-spin text-neon-cyan" />
            <span>Analyzing...</span>
          </div>
        )}
      </div>

      {/* Input area */}
      <div className="px-6 py-4 border-t border-border-default bg-bg-secondary/30">
        <div className="flex gap-2 items-end">
          <textarea
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask about your system security..."
            rows={1}
            className={cn(
              'flex-1 resize-none bg-bg-card border border-border-default rounded-lg px-4 py-3',
              'text-sm text-text-primary placeholder:text-text-muted',
              'focus:outline-none focus:border-neon-cyan/50 focus:ring-1 focus:ring-neon-cyan/20',
              'transition-colors',
            )}
            disabled={loading}
          />
          <button
            onClick={handleSend}
            disabled={loading || !input.trim()}
            className={cn(
              'flex items-center justify-center w-10 h-10 rounded-lg transition-all',
              'bg-neon-cyan/10 border border-neon-cyan/30 text-neon-cyan',
              'hover:bg-neon-cyan/20 hover:glow-cyan',
              'disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-neon-cyan/10',
            )}
          >
            <Send size={16} />
          </button>
        </div>
        <p className="text-xs text-text-muted mt-2">
          Press Enter to send, Shift+Enter for new line
        </p>
      </div>
    </div>
  );
}

function MessageBubble({ message }: { message: Message }) {
  const isUser = message.role === 'user';

  return (
    <div className={cn('flex gap-3', isUser ? 'justify-end' : 'justify-start')}>
      {!isUser && (
        <div className="flex items-start pt-1">
          <div className="w-7 h-7 rounded-full bg-neon-cyan/10 border border-neon-cyan/30 flex items-center justify-center shrink-0">
            <Bot size={14} className="text-neon-cyan" />
          </div>
        </div>
      )}

      <div className={cn(
        'max-w-[75%] rounded-xl px-4 py-3',
        isUser
          ? 'bg-neon-cyan/10 border border-neon-cyan/20 text-text-primary'
          : 'bg-bg-card border border-border-default text-text-primary',
      )}>
        {/* Tool calls rendered before content */}
        {message.toolCalls && message.toolCalls.length > 0 && (
          <div className="mb-2">
            {message.toolCalls.map((tc) => (
              <ToolCallCard key={tc.id} toolCall={tc} />
            ))}
          </div>
        )}

        {/* Message content — render markdown-like text */}
        <div className="text-sm leading-relaxed whitespace-pre-wrap">
          <FormattedContent text={message.content} />
        </div>

        <div className="text-xs text-text-muted mt-2">
          {message.timestamp.toLocaleTimeString()}
        </div>
      </div>

      {isUser && (
        <div className="flex items-start pt-1">
          <div className="w-7 h-7 rounded-full bg-bg-card border border-border-default flex items-center justify-center shrink-0">
            <User size={14} className="text-text-secondary" />
          </div>
        </div>
      )}
    </div>
  );
}

function FormattedContent({ text }: { text: string }) {
  // Simple inline formatting: **bold**, `code`, and code blocks
  const parts = text.split(/(```[\s\S]*?```|\*\*.*?\*\*|`[^`]+`)/g);

  return (
    <>
      {parts.map((part, i) => {
        if (part.startsWith('```') && part.endsWith('```')) {
          const code = part.slice(3, -3).replace(/^\w+\n/, '');
          return (
            <pre key={i} className="bg-bg-primary rounded p-2 my-2 text-xs overflow-x-auto text-neon-green">
              {code}
            </pre>
          );
        }
        if (part.startsWith('**') && part.endsWith('**')) {
          return <strong key={i} className="text-text-primary font-semibold">{part.slice(2, -2)}</strong>;
        }
        if (part.startsWith('`') && part.endsWith('`')) {
          return <code key={i} className="bg-bg-primary px-1.5 py-0.5 rounded text-xs text-neon-cyan">{part.slice(1, -1)}</code>;
        }
        return <span key={i}>{part}</span>;
      })}
    </>
  );
}

function EmptyState() {
  const suggestions = [
    'Scan my network for open ports',
    'Check for lateral movement threats',
    'Analyze the blast radius of my network',
    'Inspect loaded libraries of PID 1',
  ];

  return (
    <div className="flex flex-col items-center justify-center h-full text-center py-12">
      <div className="w-16 h-16 rounded-2xl bg-neon-cyan/5 border border-neon-cyan/20 flex items-center justify-center mb-4">
        <Bot size={28} className="text-neon-cyan" />
      </div>
      <h2 className="text-lg font-bold text-text-primary mb-1">NeuroPanopticon</h2>
      <p className="text-sm text-text-secondary mb-6 max-w-md">
        Your locally-hosted cybersecurity auditor. Ask me to scan, analyze, or remediate security issues on this machine.
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-w-lg">
        {suggestions.map((s) => (
          <button
            key={s}
            className="text-xs text-left px-3 py-2 rounded-lg border border-border-default bg-bg-card hover:bg-bg-card-hover hover:border-neon-cyan/30 text-text-secondary hover:text-text-primary transition-colors"
          >
            {s}
          </button>
        ))}
      </div>
    </div>
  );
}
