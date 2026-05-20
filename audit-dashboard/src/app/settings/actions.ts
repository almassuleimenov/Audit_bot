'use server';

import { apiFetch } from '@/lib/api';

interface SurveyQuestion {
  id: number;
  text_ru: string;
  text_kk: string;
  options_ru: string;
  options_kk: string;
  order_num: number;
  is_active: boolean;
}

export async function toggleQuestionActive(question: SurveyQuestion) {
  const updated = { ...question, is_active: !question.is_active };
  try {
    await apiFetch(`/api/questions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updated)
    });
    return { success: true, updated };
  } catch (error) {
    return { success: false, error: error instanceof Error ? error.message : 'Ошибка' };
  }
}

export async function createQuestion(payload: Omit<SurveyQuestion, 'id'> & { id: number }) {
  try {
    await apiFetch(`/api/questions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    return { success: true };
  } catch (error) {
    return { success: false, error: error instanceof Error ? error.message : 'Ошибка' };
  }
}
