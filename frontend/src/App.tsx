import { useState } from 'react';
import { Layout, type View } from '@/components/Layout';
import { Dashboard } from '@/components/Dashboard';
import { Chat } from '@/components/Chat';
import { Settings } from '@/components/Settings';

function App() {
  const [view, setView] = useState<View>('dashboard');

  return (
    <Layout currentView={view} onNavigate={setView}>
      {view === 'dashboard' && <Dashboard />}
      {view === 'chat' && <Chat />}
      {view === 'settings' && <Settings />}
    </Layout>
  );
}

export default App;
