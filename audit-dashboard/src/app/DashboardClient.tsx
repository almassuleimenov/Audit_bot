'use client';
// D:\Project\backend_projects\audit_bot\audit-dashboard\src\app\DashboardClient.tsx
import { useEffect, useState } from 'react';
import { StatsData } from './page';
// Строгая типизация
interface LeadEvent {
  id?: string;
  phone: string;
  username?: string;
  status: string;
}

interface DashboardClientProps {
  stats: StatsData; // Обязательное описание пропса
}
export default function DashboardClient({ stats }: DashboardClientProps) {
  const [leads, setLeads] = useState<LeadEvent[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting');
  const [errorMessage, setErrorMessage] = useState<string>('');

  useEffect(() => {
    // Внимание: при настроенном прокси (Caddy) и включенном JWT, 
    // браузер автоматически прикрепит HTTP-Only куки к этому GET-запросу.
    const eventSource = new EventSource('/api/stream/leads');

    eventSource.onopen = () => {
      setConnectionStatus('connected');
      setErrorMessage('');
    };

    eventSource.onmessage = (event) => {
      try {
        const newLead: LeadEvent = JSON.parse(event.data);
        // O(1) добавление в начало массива
        setLeads((prev) => [newLead, ...prev]); 
      } catch (error) {
        console.error('SSE parsing error:', error);
        setErrorMessage(`Parse error: ${error instanceof Error ? error.message : 'Unknown error'}`);
      }
    };

    eventSource.onerror = () => {
      setConnectionStatus('disconnected');
      
      // Получаем детальное объяснение ошибки на основе readyState
      let errorMsg = 'Connection closed';
      if (eventSource.readyState === EventSource.CONNECTING) {
        errorMsg = 'Connecting... (pending)';
      } else if (eventSource.readyState === EventSource.CLOSED) {
        errorMsg = 'Connection closed. Possible causes: 401 Unauthorized (need login), network error, or server error. Check browser console for details.';
      }
      
      console.error('[SSE Error]', {
        readyState: eventSource.readyState,
        url: eventSource.url,
        timestamp: new Date().toISOString()
      });
      
      setErrorMessage(errorMsg);
    };

    return () => {
      eventSource.close();
    };
  }, []);

  return (
    <div className="space-y-6">
      {/* Статус соединения */}
      <div className="flex items-center space-x-2 p-4 bg-neutral-900 border border-neutral-800 rounded-xl">
        <span className={`h-2.5 w-2.5 rounded-full ${
          connectionStatus === 'connected' ? 'bg-emerald-500 animate-pulse' : 
          connectionStatus === 'connecting' ? 'bg-amber-500' : 'bg-rose-500'
        }`} />
        <span className="text-xs font-mono uppercase tracking-wider text-neutral-400">
          Stream: {connectionStatus}
        </span>
        {errorMessage && (
          <span className="text-xs font-mono text-rose-400 ml-auto truncate" title={errorMessage}>
            {errorMessage}
          </span>
        )}
      </div>

      {/* Поток лидов */}
      <div className="grid grid-cols-1 gap-6">
        <div className="bg-neutral-900 border border-neutral-800 rounded-2xl p-6">
          <h3 className="text-lg font-medium text-neutral-100 mb-4 font-sans tracking-tight">Регистрации (Real-time)</h3>
          
          {leads.length === 0 ? (
            <div className="flex items-center justify-center h-48 border border-dashed border-neutral-800 rounded-xl">
              <p className="text-sm text-neutral-500 font-mono">Ожидание событий...</p>
            </div>
          ) : (
            <div className="space-y-3 max-h-[500px] overflow-y-auto pr-2">
              {leads.map((lead) => (
                <div key={lead.id || lead.phone} className="flex items-center justify-between p-4 bg-neutral-950 border border-neutral-800 rounded-xl">
                  <span className="text-sm font-mono text-neutral-200">{lead.phone}</span>
                  <span className="text-xs px-2.5 py-1 font-mono font-semibold uppercase bg-emerald-500/10 text-emerald-400 rounded-md border border-emerald-500/20">
                    {lead.status}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}