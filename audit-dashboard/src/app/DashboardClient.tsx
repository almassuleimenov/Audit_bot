'use client';
// D:\Project\backend_projects\audit_bot\audit-dashboard\src\app\DashboardClient.tsx

import { useEffect, useState } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { StatsData } from './page';

interface LeadEvent {
  id?: string;
  phone: string;
  username?: string;
  status: string;
}

interface DashboardClientProps {
  stats: StatsData;
}

export default function DashboardClient({ stats }: DashboardClientProps) {
  const [leads, setLeads] = useState<LeadEvent[]>([]);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting');
  const [errorMessage, setErrorMessage] = useState<string>('');

  useEffect(() => {
    const eventSource = new EventSource('https://audit-bot-bok1.onrender.com/api/stream/leads', {
      withCredentials: true 
    });

    eventSource.onopen = () => {
      console.log('[SRE] SSE Connection established');
      setConnectionStatus('connected');
      setErrorMessage('');
    };

    eventSource.onmessage = (event) => {
      try {
        const newLead: LeadEvent = JSON.parse(event.data);
        // O(1) добавление в начало массива
        setLeads((prev) => [newLead, ...prev]); 
      } catch (error) {
        console.error('[SRE] SSE parsing error:', error);
        setErrorMessage(`Parse error: ${error instanceof Error ? error.message : 'Unknown error'}`);
      }
    };

    eventSource.onerror = (error) => {
      setConnectionStatus('disconnected');
      
      let errorMsg = 'Connection closed';
      if (eventSource.readyState === EventSource.CONNECTING) {
        errorMsg = 'Connecting... (pending)';
      } else if (eventSource.readyState === EventSource.CLOSED) {
        errorMsg = 'Connection closed. Possible causes: 401 Unauthorized, network error, or server crash.';
      }
      
      console.error('[SRE] SSE Stream disconnected:', {
        readyState: eventSource.readyState,
        url: eventSource.url,
        timestamp: new Date().toISOString()
      });
      
      setErrorMessage(errorMsg);
      eventSource.close();
    };

    // Strict Rule: Return всегда в самом конце блока
    return () => {
      eventSource.close();
    };
  }, []);

  return (
    <div className="space-y-6">
      {/* Секция статистики (Верхний ряд) */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-6 shadow-sm">
          <p className="text-sm text-neutral-400 font-mono tracking-wide uppercase">Всего аудитов</p>
          <p className="text-3xl font-semibold text-neutral-100 mt-2">{stats.total_audits || 0}</p>
        </div>
        <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-6 shadow-sm">
          <p className="text-sm text-neutral-400 font-mono tracking-wide uppercase">Средний балл</p>
          <p className="text-3xl font-semibold text-emerald-400 mt-2">
            {stats.average_score ? stats.average_score.toFixed(1) : '0.0'}
          </p>
        </div>
        <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-6 shadow-sm">
          <p className="text-sm text-neutral-400 font-mono tracking-wide uppercase">Обращения (Лиды)</p>
          <p className="text-3xl font-semibold text-neutral-100 mt-2">{stats.total_appointments || 0}</p>
        </div>
      </div>

      {/* Секция динамики проверок (График) */}
      <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-6 shadow-sm">
        <h3 className="text-lg font-medium text-neutral-100 mb-6 font-sans tracking-tight">Динамика проверок</h3>
        {stats.daily_dynamics && stats.daily_dynamics.length > 0 ? (
          <div className="h-72 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={stats.daily_dynamics} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#262626" vertical={false} />
                <XAxis 
                  dataKey="name" 
                  stroke="#525252" 
                  fontSize={12} 
                  tickLine={false} 
                  axisLine={false} 
                />
                <YAxis 
                  stroke="#525252" 
                  fontSize={12} 
                  tickLine={false} 
                  axisLine={false} 
                  allowDecimals={false}
                />
                <Tooltip 
                  contentStyle={{ backgroundColor: '#171717', borderColor: '#262626', borderRadius: '8px' }}
                  itemStyle={{ color: '#10b981' }}
                />
                <Area 
                  type="monotone" 
                  dataKey="Проверки" 
                  stroke="#10b981" 
                  strokeWidth={2}
                  fillOpacity={1} 
                  fill="url(#colorCount)" 
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <div className="flex items-center justify-center h-72 border border-dashed border-neutral-800 rounded-xl">
            <p className="text-sm text-neutral-500 font-mono">Недостаточно данных для графика</p>
          </div>
        )}
      </div>

      {/* Статус соединения SSE */}
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

      {/* Поток лидов (Real-time) */}
      <div className="grid grid-cols-1 gap-6">
        <div className="bg-neutral-900 border border-neutral-800 rounded-xl p-6 shadow-sm">
          <h3 className="text-lg font-medium text-neutral-100 mb-4 font-sans tracking-tight">Регистрации (Real-time)</h3>
          
          {leads.length === 0 ? (
            <div className="flex items-center justify-center h-48 border border-dashed border-neutral-800 rounded-xl">
              <p className="text-sm text-neutral-500 font-mono">Ожидание событий...</p>
            </div>
          ) : (
            <div className="space-y-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar">
              {leads.map((lead, idx) => (
                <div key={lead.id || idx} className="flex items-center justify-between p-4 bg-neutral-950 border border-neutral-800 rounded-xl hover:border-neutral-700 transition-colors">
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