'use client';
// src/app/DashboardClient.tsx
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';

export default function DashboardClient({ stats }: { stats: any }) {
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

        {/* ... (остальной твой код с графиками LineChart и PieChart оставляешь без изменений) ... */}
        
      </div>
    </main>
  );
}