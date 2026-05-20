'use client';

import { useState } from 'react';
import { toggleQuestionActive, createQuestion } from './actions';

interface SurveyQuestion {
  id: number;
  text_ru: string;
  text_kk: string;
  options_ru: string;
  options_kk: string;
  order_num: number;
  is_active: boolean;
}

interface SettingsClientProps {
  initialQuestions: SurveyQuestion[];
}

export default function SettingsClient({ initialQuestions }: SettingsClientProps) {
  const [questions, setQuestions] = useState<SurveyQuestion[]>(initialQuestions);
  const [isAdding, setIsAdding] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [newQ, setNewQ] = useState({
    text_ru: '',
    text_kk: '',
    options_ru: '["Да", "Нет", "Затрудняюсь"]',
    options_kk: '["Иә", "Жоқ", "Қиналамын"]'
  });

  const toggleActive = async (q: SurveyQuestion) => {
    setIsLoading(true);
    try {
      const result = await toggleQuestionActive(q);
      if (result.success && result.updated) {
        setQuestions(questions.map(item => item.id === q.id ? result.updated : item));
      } else {
        alert("Ошибка при сохранении");
      }
    } catch (e) {
      alert("Ошибка при сохранении");
    } finally {
      setIsLoading(false);
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

    const payload = {
      id: 0,
      text_ru: newQ.text_ru,
      text_kk: newQ.text_kk,
      options_ru: newQ.options_ru,
      options_kk: newQ.options_kk,
      order_num: nextOrderNum,
      is_active: true
    };

    setIsLoading(true);
    try {
      const result = await createQuestion(payload);
      if (result.success) {
        setIsAdding(false);
        setNewQ({
          text_ru: '',
          text_kk: '',
          options_ru: '["Да", "Нет", "Затрудняюсь"]',
          options_kk: '["Иә", "Жоқ", "Қиналамын"]'
        });
        // Перезагрузить страницу для получения новых данных
        location.reload();
      } else {
        alert("Ошибка при создании вопроса");
      }
    } catch (e) {
      alert("Ошибка при создании вопроса");
    } finally {
      setIsLoading(false);
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
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-3xl font-bold text-[#182442] mb-2">Вопросы опроса</h1>
            <p className="text-gray-500">Управление вопросами и вариантами ответов для бота</p>
          </div>
          <button
            onClick={() => setIsAdding(!isAdding)}
            disabled={isLoading}
            className="px-6 py-2 bg-[#296956] text-white rounded-xl hover:bg-[#1f523e] disabled:opacity-50 transition-colors font-medium"
          >
            {isAdding ? 'Отмена' : '+ Добавить вопрос'}
          </button>
        </div>

        {isAdding && (
          <div className="bg-white p-6 rounded-2xl border border-gray-200 shadow-sm">
            <h3 className="text-lg font-bold text-[#182442] mb-4">Новый вопрос</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Вопрос на русском</label>
                <input
                  type="text"
                  value={newQ.text_ru}
                  onChange={(e) => setNewQ({ ...newQ, text_ru: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#296956]"
                  placeholder="Введите вопрос"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Вопрос на казахском</label>
                <input
                  type="text"
                  value={newQ.text_kk}
                  onChange={(e) => setNewQ({ ...newQ, text_kk: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#296956]"
                  placeholder="Введите вопрос"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Варианты (русский, JSON)</label>
                <textarea
                  value={newQ.options_ru}
                  onChange={(e) => setNewQ({ ...newQ, options_ru: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#296956]"
                  rows={3}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Варианты (казахский, JSON)</label>
                <textarea
                  value={newQ.options_kk}
                  onChange={(e) => setNewQ({ ...newQ, options_kk: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#296956]"
                  rows={3}
                />
              </div>
            </div>
            <button
              onClick={handleAddQuestion}
              disabled={isLoading}
              className="mt-4 w-full px-4 py-2 bg-[#296956] text-white rounded-lg hover:bg-[#1f523e] disabled:opacity-50 transition-colors font-medium"
            >
              {isLoading ? 'Сохранение...' : 'Добавить вопрос'}
            </button>
          </div>
        )}

        <div className="space-y-3">
          {questions.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <p>Вопросов не найдено</p>
            </div>
          ) : (
            questions.map((question) => (
              <div key={question.id} className="bg-white p-4 rounded-xl border border-gray-200 shadow-sm hover:shadow-md transition-shadow">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <h3 className="font-medium text-[#182442] mb-1">
                      Q{question.order_num}: {question.text_ru}
                    </h3>
                    <p className="text-sm text-gray-600 mb-2">{question.text_kk}</p>
                    <div className="flex gap-4 text-sm">
                      <span className="text-gray-500">
                        РУ: {renderOptions(question.options_ru)}
                      </span>
                      <span className="text-gray-500">
                        КК: {renderOptions(question.options_kk)}
                      </span>
                    </div>
                  </div>
                  <button
                    onClick={() => toggleActive(question)}
                    disabled={isLoading}
                    className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                      question.is_active
                        ? 'bg-[#aff0d8] text-[#296956] hover:bg-[#90e8c2] disabled:opacity-50'
                        : 'bg-gray-200 text-gray-700 hover:bg-gray-300 disabled:opacity-50'
                    }`}
                  >
                    {question.is_active ? 'Активен' : 'Неактивен'}
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </main>
  );
}
