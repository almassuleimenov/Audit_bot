// src/app/page.tsx
'use client';

import { useEffect, useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';

interface AuditRecord {
  ID: number;
  Score: number;
  CreatedAt: string;
}

interface Appointment {
  ID: number;
  CreatedAt: string;
}

export default function OverviewPage() {
  const [audits, setAudits] = useState<AuditRecord[]>([]);
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Скачиваем данные при загрузке страницы
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [resAudits, resAppointments] = await Promise.all([
          fetch('https://audit-bot-bok1.onrender.com/api/audits'),
          fetch('https://audit-bot-bok1.onrender.com/api/appointments')
        ]);
        
        const auditsData = await resAudits.json();
        const apptsData = await resAppointments.json();
        
        setAudits(auditsData || []);
        setAppointments(apptsData || []);
      } catch (error) {
        console.error("Ошибка загрузки данных:", error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  // Высчитываем статистику
  const totalAudits = audits.length;
  const totalAppointments = appointments.length;
  const averageScore = totalAudits > 0 
    ? (audits.reduce((acc, curr) => acc + curr.Score, 0) / totalAudits).toFixed(1)
    : "0.0";

  // Подготавливаем данные для кругового графика (Распределение оценок)
  let score5 = 0, score4 = 0, score1to3 = 0;
  audits.forEach(a => {
    if (a.Score === 5) score5++;
    else if (a.Score === 4) score4++;
    else score1to3++;
  });

  const donutData = [
    { name: 'Оценка 5', value: score5, fill: '#aff0d8' }, // Зеленый
    { name: 'Оценка 4', value: score4, fill: '#bac6ec' }, // Синий
    { name: 'Оценки 1-3', value: score1to3, fill: '#ffb4a9' } // Красный
  ];

  // Подготавливаем данные для линейного графика (Динамика по дням)
  const lineDataMap: Record<string, number> = {};
  audits.forEach(a => {
    const date = new Date(a.CreatedAt).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });
    lineDataMap[date] = (lineDataMap[date] || 0) + 1;
  });
  const lineData = Object.keys(lineDataMap).map(key => ({ name: key, Проверки: lineDataMap[key] }));

  if (isLoading) {
    return <div className="p-8 text-center text-gray-500">Загрузка панели управления...</div>;
  }

  return (
    <main className="flex flex-col min-h-screen">
      <header className="sticky top-0 z-50 flex justify-between items-center px-8 h-16 bg-white/70 backdrop-blur-md border-b border-gray-200">
        <h2 className="font-bold text-[#182442] text-xl">Панель управления</h2>
      </header>

      <div className="p-8 max-w-7xl mx-auto w-full flex flex-col gap-6">
        <div>
          <h2 className="text-3xl font-bold text-[#182442] mb-1">Доброе утро!</h2>
          <p className="text-gray-500">Вот краткий обзор текущей ситуации на сегодня.</p>
        </div>

        {/* Карточки со статистикой */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm">
            <p className="text-gray-500 mb-1">Пройденных анкет</p>
            <h3 className="text-4xl font-bold text-[#182442]">{totalAudits}</h3>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm">
            <p className="text-gray-500 mb-1">Средний балл аудиторов</p>
            <div className="flex items-baseline gap-2">
              <h3 className="text-4xl font-bold text-[#182442]">{averageScore}</h3>
              <span className="text-gray-400">/ 5.0</span>
            </div>
          </div>
          <div className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm">
            <p className="text-gray-500 mb-1">Заявок на прием</p>
            <h3 className="text-4xl font-bold text-[#182442]">{totalAppointments}</h3>
          </div>
        </div>

        {/* Графики */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Линейный график */}
          {/* Линейный график */}
<div className="lg:col-span-2 bg-white rounded-2xl p-6 border border-gray-100 shadow-sm h-[400px] flex flex-col">
  <h3 className="font-bold text-[#182442] mb-6">Динамика поступления анкет</h3>
  {/* ДОБАВЛЕН STYLE СЮДА 👇 */}
  <div className="flex-1 w-full h-full" style={{ minHeight: '300px' }}>
    <ResponsiveContainer width="100%" height="100%">
      <LineChart data={lineData}>
        {/* ... содержимое графика ... */}
      </LineChart>
    </ResponsiveContainer>
  </div>
</div>

          {/* Круговой график */}
          <div className="bg-white rounded-2xl p-6 border border-gray-100 shadow-sm h-[400px] flex flex-col">
  <h3 className="font-bold text-[#182442] mb-2">Распределение оценок</h3>
  {/* ДОБАВЛЕН STYLE СЮДА 👇 */}
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
              {/* Легенда под графиком */}
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