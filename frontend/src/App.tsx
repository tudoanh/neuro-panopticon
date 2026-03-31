import { useState } from 'react';
import { Layout, type View } from '@/components/Layout';
import { Dashboard } from '@/components/Dashboard';
import { Chat } from '@/components/Chat';

function App() {
  const [view, setView] = useState<View>('dashboard');

  return (
    <Layout currentView={view} onNavigate={setView}>
      {view === 'dashboard' && <Dashboard />}
      {view === 'chat' && <Chat />}
      {view === 'settings' && <PlaceholderView title="Settings" description="LLM backend configuration coming soon." />}
    </Layout>
  );
}

function PlaceholderView({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex items-center justify-center h-full">
      <div className="text-center">
        <h2 className="text-lg font-bold text-text-secondary">{title}</h2>
        <p className="text-sm text-text-muted mt-2">{description}</p>
      </div>
    </div>
  );
}

export default App;
