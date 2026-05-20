import { apiFetch } from '@/lib/api';
import DashboardClient from './DashboardClient';
// src/app/page.tsx
interface StatsData {
  total_audits: number;
  average_score: number;
  total_appointments: number;
  score_distribution: Record<string, number>;
  daily_dynamics: { name: string; Проверки: number }[];
}

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

  return <DashboardClient stats={stats} />;
}