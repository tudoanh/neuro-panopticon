import { useState } from 'react';
import { Shield, MessageSquare, Settings, Activity } from 'lucide-react';
import { cn } from '@/lib/utils';

export type View = 'dashboard' | 'chat' | 'settings';

interface LayoutProps {
  currentView: View;
  onNavigate: (view: View) => void;
  children: React.ReactNode;
}

const navItems: { id: View; label: string; icon: React.ReactNode }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: <Shield size={20} /> },
  { id: 'chat', label: 'Chat', icon: <MessageSquare size={20} /> },
  { id: 'settings', label: 'Settings', icon: <Settings size={20} /> },
];

export function Layout({ currentView, onNavigate, children }: LayoutProps) {
  return (
    <div className="flex h-screen bg-bg-primary">
      {/* Sidebar */}
      <nav className="flex flex-col w-16 bg-bg-secondary border-r border-border-default">
        {/* Logo */}
        <div className="flex items-center justify-center h-16 border-b border-border-default">
          <Activity size={24} className="text-neon-cyan" />
        </div>

        {/* Nav items */}
        <div className="flex flex-col items-center gap-2 py-4 flex-1">
          {navItems.map((item) => (
            <button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              className={cn(
                'flex items-center justify-center w-10 h-10 rounded-lg transition-all duration-200',
                'hover:bg-bg-card-hover',
                currentView === item.id
                  ? 'bg-bg-card text-neon-cyan glow-cyan'
                  : 'text-text-secondary hover:text-text-primary'
              )}
              title={item.label}
            >
              {item.icon}
            </button>
          ))}
        </div>

        {/* Status indicator */}
        <div className="flex items-center justify-center py-4 border-t border-border-default">
          <div className="w-2 h-2 rounded-full bg-neon-green animate-pulse" title="Scanner active" />
        </div>
      </nav>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        {children}
      </main>
    </div>
  );
}
