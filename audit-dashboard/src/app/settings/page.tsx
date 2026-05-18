// src/app/settings/page.tsx
'use client';

import { useEffect, useState } from 'react';

interface SurveyQuestion {
  id: number;
  text_ru: string;
  text_kk: string;
  options_ru: string; // JSON строка
  options_kk: string; // JSON строка
  order_num: number;
  is_active: boolean;
}

export default function SettingsPage() {
  const [questions, setQuestions] = useState<SurveyQuestion[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchQuestions = async () => {
    try {
      const res = await fetch('https://audit-bot-bok1.onrender.com/api/questions');
      const data = await res.json();
      setQuestions(data || []);
    } catch (error) {
      console.error("Ошибка загрузки:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchQuestions();
  }, []);

  const toggleActive = async (q: SurveyQuestion) => {
    const updated = { ...q, is_active: !q.is_active };
    try {
      await fetch('https://audit-bot-bok1.onrender.com/api/questions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updated)
      });
      fetchQuestions(); // Обновляем список
    } catch (e) {
      alert("Ошибка при сохранении");
    }
  };

  // Вспомогательная функция для красивого отображения вариантов
  const renderOptions = (jsonString: string) => {
    try {
      const arr = JSON.parse(jsonString);
      return arr.join(' / ');
    } catch {
      return jsonString;
    }
  };

  return (
    <main className="flex flex-col min-h-screen">
      <header className="sticky top-0 z-50 flex items-center px-8 h-16 bg-white/70 backdrop-blur-md border-b border-gray-200">
        <h2 className="font-bold text-[#182442] text-xl">Управление вопросами бота</h2>
      </header>

      <div className="p-8 max-w-7xl mx-auto w-full flex flex-col gap-8">
        <div>
          <h2 className="text-3xl font-bold text-[#182442] mb-2">Настройки анкеты</h2>
          <p className="text-gray-500">Здесь можно включать и отключать вопросы. (Редактирование текста можно добавить по клику)</p>
        </div>

        {isLoading ? (
          <p className="text-gray-500">Загрузка данных...</p>
        ) : (
          <div className="flex flex-col gap-4">
            {questions.map((q) => (
              <div key={q.id} className={`bg-white rounded-[24px] p-6 shadow-sm border ${q.is_active ? 'border-l-4 border-l-[#aff0d8]' : 'border-l-4 border-l-gray-300 opacity-60'}`}>
                <div className="flex justify-between items-start mb-4">
                  <h3 className="text-lg font-bold text-[#182442]">Вопрос {q.order_num}</h3>
                  <button 
                    onClick={() => toggleActive(q)}
                    className={`px-4 py-2 rounded-xl text-sm font-bold ${q.is_active ? 'bg-red-100 text-red-600 hover:bg-red-200' : 'bg-green-100 text-green-600 hover:bg-green-200'}`}
                  >
                    {q.is_active ? 'Отключить' : 'Включить'}
                  </button>
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="bg-gray-50 p-4 rounded-xl border border-gray-200">
                    <span className="text-gray-500 text-xs block mb-1">Русский язык</span>
                    <p className="font-medium text-sm text-[#182442] mb-3">{q.text_ru}</p>
                    <span className="text-gray-500 text-xs block mb-1">Кнопки:</span>
                    <p className="text-sm font-bold text-[#296956]">{renderOptions(q.options_ru)}</p>
                  </div>

                  <div className="bg-gray-50 p-4 rounded-xl border border-gray-200">
                    <span className="text-gray-500 text-xs block mb-1">Казахский язык</span>
                    <p className="font-medium text-sm text-[#182442] mb-3">{q.text_kk}</p>
                    <span className="text-gray-500 text-xs block mb-1">Кнопки:</span>
                    <p className="text-sm font-bold text-[#296956]">{renderOptions(q.options_kk)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </main>
  );
}