'use client';

import { useEffect, useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';

interface StatsData {
  total_audits: number;
  average_score: number;
  total_appointments: number;
  score_distribution: Record<string, number>;
  daily_dynamics: { name: string; Проверки: number }[];
}

export default function OverviewPage() {
  const [stats, setStats] = useState<StatsData | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const res = await fetch('https://audit-bot-bok1.onrender.com/api/stats');
        const data = await res.json();
        setStats(data);
      } catch (error) {
        console.error("Ошибка загрузки аналитики:", error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchStats();
  }, []);

  if (isLoading || !stats) {
    return <div className="p-8 text-center text-gray-500">Загрузка панели управления...</div>;
  }

  // Данные для кругового графика напрямую из API
  const donutData = [
    { name: 'Оценка 5', value: stats.score_distribution["5"] || 0, fill: '#aff0d8' },
    { name: 'Оценка 4', value: stats.score_distribution["4"] || 0, fill: '#bac6ec' },
    { name: 'Оценки 1-3', value: (stats.score_distribution["3"] || 0) + (stats.score_distribution["2"] || 0) + (stats.score_distribution["1"] || 0), fill: '#ffb4a9' }
  ];

  return (
    <main className="flex flex-col min-h-screen">
      <header className="sticky top-0 z-50 flex justify-between items-center px-8 h-16 bg-white/70 backdrop-blur-md border-b border-gray-200">
        <h2 className="font-bold text-[#182442] text-xl">Панель управления</h2>
      </header>

      <div className="p-8 max-w-7xl mx-auto w-full flex flex-col gap-6">
        <div>
          <h2 className="text-3xl font-bold text-[#182442] mb-1">Доброе утро!</h2>
          <p className="text-gray-500">Вот краткий обзор текущей ситуации.</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm">
            <p className="text-gray-500 mb-1">Пройденных анкет</p>
            <h3 className="text-4xl font-bold text-[#182442]">{stats.total_audits}</h3>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm">
            <p className="text-gray-500 mb-1">Средний балл аудиторов</p>
            <div className="flex items-baseline gap-2">
              <h3 className="text-4xl font-bold text-[#182442]">{stats.average_score.toFixed(1)}</h3>
              <span className="text-gray-400">/ 5.0</span>
            </div>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm">
            <p className="text-gray-500 mb-1">Заявок на прием</p>
            <h3 className="text-4xl font-bold text-[#182442]">{stats.total_appointments}</h3>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 bg-white rounded-2xl p-6 border border-gray-100 shadow-sm h-[400px] flex flex-col">
            <h3 className="font-bold text-[#182442] mb-6">Динамика поступления анкет</h3>
            <div className="flex-1 w-full h-full" style={{ minHeight: '300px' }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={stats.daily_dynamics}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f0f0f0" />
                  <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{fill: '#8898aa', fontSize: 12}} dy={10} />
                  <YAxis axisLine={false} tickLine={false} tick={{fill: '#8898aa', fontSize: 12}} dx={-10} />
                  <Tooltip 
                    contentStyle={{borderRadius: '12px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)'}}
                    itemStyle={{color: '#182442', fontWeight: 600}}
                  />
                  <Line type="monotone" dataKey="Проверки" stroke="#296956" strokeWidth={3} dot={{r: 4, fill: '#296956', strokeWidth: 0}} activeDot={{r: 6, strokeWidth: 0}} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm h-[400px] flex flex-col">
            <h3 className="font-bold text-[#182442] mb-2">Распределение оценок</h3>
            <div className="flex-1 w-full h-full relative" style={{ minHeight: '300px' }}>
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={donutData} innerRadius={60} outerRadius={80} paddingAngle={5} dataKey="value" stroke="none">
                    {donutData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.fill} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
              <div className="absolute bottom-0 w-full flex flex-col gap-2">
                {donutData.map((item, idx) => (
                  <div key={idx} className="flex justify-between items-center text-sm">
                    <div className="flex items-center gap-2">
                      <div className="w-3 h-3 rounded-full" style={{backgroundColor: item.fill}}></div>
                      <span className="text-gray-600">{item.name}</span>
                    </div>
                    <span className="font-bold text-[#182442]">{item.value}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}