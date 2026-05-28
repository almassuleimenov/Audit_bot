import { apiFetch } from '@/lib/api';
import DashboardClient from './DashboardClient';
// src/app/page.tsx

// Добавь ключевое слово export
export interface StatsData {
  total_audits: number;
  average_score: number;
  total_appointments: number;
  score_distribution: Record<string, number>;
  daily_dynamics: { name: string; "Проверки": number }[];
}

// ... остальной код страницы без изменений

export default async function OverviewPage() {
  let stats: StatsData | null = null;
  let error: string | null = null;

  
  try {
    stats = await apiFetch<StatsData>('/api/stats');
  } catch (err) {
    error = err instanceof Error ? err.message : 'Ошибка загрузки аналитики';
  }

  if (error) {
    return (
      <div className="p-8 text-center text-red-500">
        <p>Ошибка загрузки панели управления: {error}</p>
      </div>
    );
  }

  if (!stats) {
    return <div className="p-8 text-center text-gray-500">Нет данных для отображения</div>;
  }

  return       <div className="p-8 max-w-7xl mx-auto w-full flex flex-col gap-8">
        <div>
          <h2 className="text-3xl font-bold text-[#182442] mb-2">Обзор аналитики</h2>
          <p className="text-gray-500">Ключевые показатели и динамика проверок</p>
        </div>
        <DashboardClient stats={stats} />
      </div>;
}