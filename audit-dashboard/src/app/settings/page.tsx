export const dynamic = "force-dynamic";
import { apiFetch } from '@/lib/api';
import SettingsClient from './SettingsClient';

interface SurveyQuestion {
  id: number;
  text_ru: string;
  text_kk: string;
  options_ru: string;
  options_kk: string;
  order_num: number;
  is_active: boolean;
}

export default async function SettingsPage() {
  let questions: SurveyQuestion[] = [];
  let error: string | null = null;

  try {
    questions = await apiFetch<SurveyQuestion[]>('/api/questions');
  } catch (err) {
    error = err instanceof Error ? err.message : 'Ошибка загрузки данных';
    console.error("Ошибка загрузки вопросов:", err);
  }

  if (error) {
    return (
      <div className="p-8 text-center text-red-500">
        <p>Ошибка загрузки данных: {error}</p>
      </div>
    );
  }

  return <SettingsClient initialQuestions={questions} />;
}