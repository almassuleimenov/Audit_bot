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
  
  // Состояния для формы создания
  const [isAdding, setIsAdding] = useState(false);
  const [newQ, setNewQ] = useState({
    text_ru: '',
    text_kk: '',
    options_ru: '["Да", "Нет", "Затрудняюсь"]',
    options_kk: '["Иә", "Жоқ", "Қиналамын"]'
  });

  const fetchQuestions = async () => {
    try {
      const res = await fetch('https://audit-bot-1-j1xj.onrender.com/api/questions');
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
      await fetch('https://audit-bot-1-j1xj.onrender.com/api/questions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updated)
      });
      fetchQuestions();
    } catch (e) {
      alert("Ошибка при сохранении");
    }
  };

  const handleAddQuestion = async () => {
    if (!newQ.text_ru.trim() || !newQ.text_kk.trim()) {
      alert("Пожалуйста, заполните текст вопроса на обоих языках.");
      return;
    }

    try {
      JSON.parse(newQ.options_ru);
      JSON.parse(newQ.options_kk);
    } catch (e) {
      alert("Варианты ответов должны быть в формате JSON массива. Пример: [\"Да\", \"Нет\"]");
      return;
    }

    const nextOrderNum = questions.length > 0 
      ? Math.max(...questions.map(q => q.order_num)) + 1 
      : 1;

    const payload: SurveyQuestion = {
      id: 0, // 0 заставит GORM создать новую запись
      text_ru: newQ.text_ru,
      text_kk: newQ.text_kk,
      options_ru: newQ.options_ru,
      options_kk: newQ.options_kk,
      order_num: nextOrderNum,
      is_active: true
    };

    try {
      await fetch('https://audit-bot-1-j1xj.onrender.com/api/questions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      setIsAdding(false);
      setNewQ({
        text_ru: '',
        text_kk: '',
        options_ru: '["Да", "Нет", "Затрудняюсь"]',
        options_kk: '["Иә", "Жоқ", "Қиналамын"]'
      });
      fetchQuestions();
    } catch (e) {
      alert("Ошибка при создании вопроса");
    }
  };

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
        <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-4">
          <div>
            <h2 className="text-3xl font-bold text-[#182442] mb-2">Настройки анкеты</h2>
            <p className="text-gray-500">Управление вопросами, которые задает бот при аудите.</p>
          </div>
          <button 
            onClick={() => setIsAdding(!isAdding)}
            className="px-6 py-3 bg-[#182442] text-white rounded-xl font-medium hover:bg-[#2a3a63] transition-colors shadow-sm"
          >
            {isAdding ? 'Отменить' : '+ Добавить вопрос'}
          </button>
        </div>

        {/* Форма добавления */}
        {isAdding && (
          <div className="bg-white rounded-[24px] p-6 shadow-sm border border-gray-200 flex flex-col gap-4 animate-fade-in">
            <h3 className="text-xl font-bold text-[#182442] border-b border-gray-100 pb-2">Новый вопрос</h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Текст вопроса (Русский)</label>
                <textarea 
                  value={newQ.text_ru}
                  onChange={(e) => setNewQ({...newQ, text_ru: e.target.value})}
                  className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#296956]"
                  rows={2}
                  placeholder="Например: Были ли аудиторы вежливы?"
                />
                <label className="text-sm font-medium text-gray-700 mt-2">Кнопки ответов (Русский - JSON)</label>
                <input 
                  value={newQ.options_ru}
                  onChange={(e) => setNewQ({...newQ, options_ru: e.target.value})}
                  className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#296956] font-mono text-sm"
                />
              </div>

              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-gray-700">Текст вопроса (Казахский)</label>
                <textarea 
                  value={newQ.text_kk}
                  onChange={(e) => setNewQ({...newQ, text_kk: e.target.value})}
                  className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#296956]"
                  rows={2}
                  placeholder="Мысалы: Аудиторлар сыпайы болды ма?"
                />
                <label className="text-sm font-medium text-gray-700 mt-2">Кнопки ответов (Казахский - JSON)</label>
                <input 
                  value={newQ.options_kk}
                  onChange={(e) => setNewQ({...newQ, options_kk: e.target.value})}
                  className="w-full p-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#296956] font-mono text-sm"
                />
              </div>
            </div>

            <div className="flex justify-end mt-4 pt-4 border-t border-gray-100">
              <button 
                onClick={handleAddQuestion}
                className="px-6 py-3 bg-[#296956] text-white rounded-xl font-bold hover:bg-[#1f4e40] transition-colors"
              >
                Сохранить вопрос
              </button>
            </div>
          </div>
        )}

        {/* Список вопросов */}
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
                    className={`px-4 py-2 rounded-xl text-sm font-bold transition-colors ${q.is_active ? 'bg-red-100 text-red-600 hover:bg-red-200' : 'bg-green-100 text-green-600 hover:bg-green-200'}`}
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