'use client';

import { useEffect, useState } from 'react';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

// Типизация данных с твоего бэкенда
interface AuditRecord {
  ID: number;
  TelegramID: number;
  BIN: string;
  Position: string;
  Answers: string | Record<string, string>;
  Score: number;
  CreatedAt: string;
}

export default function AuditPage() {
  const [records, setRecords] = useState<AuditRecord[]>([]);
  const [selectedRecord, setSelectedRecord] = useState<AuditRecord | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Забираем данные с твоего сервера
  useEffect(() => {
    const fetchAudits = async () => {
      try {
        const res = await fetch('https://audit-bot-bok1.onrender.com/api/audits');
        if (!res.ok) throw new Error('Network response was not ok');
        const data = await res.json();
        setRecords(data || []);
        if (data && data.length > 0) {
          setSelectedRecord(data[0]); // Выбираем первую запись по умолчанию
        }
      } catch (error) {
        console.error("Failed to fetch audits:", error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchAudits();
  }, []);

  const handleExport = () => {
    window.open('https://audit-bot-bok1.onrender.com/api/export', '_blank');
  };

  const getParsedAnswers = (answers: string | Record<string, string>) => {
    if (typeof answers === 'string') {
      try { return JSON.parse(answers); } catch { return {}; }
    }
    return answers || {};
  };

  return (
    <div className="flex flex-col min-h-screen">
      {/* Шапка */}
      <header className="sticky top-0 z-50 flex justify-between items-center px-8 h-16 w-full bg-white/70 backdrop-blur-md border-b border-gray-200">
        <h2 className="font-bold text-[#182442] text-xl">Панель управления</h2>
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-full bg-gray-300 overflow-hidden border-2 border-white">
            <div className="w-full h-full bg-blue-500"></div>
          </div>
        </div>
      </header>

      {/* Контент страницы */}
      <div className="p-8 flex-1 flex flex-col gap-6 max-w-7xl mx-auto w-full">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-4 mb-4">
          <div>
            <h2 className="text-3xl font-bold text-[#182442] mb-2">Анкетирование</h2>
            <p className="text-gray-500">Просмотр обратной связи и результатов аудита.</p>
          </div>
          <button 
            onClick={handleExport}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 rounded-xl text-gray-700 hover:bg-gray-50 transition-colors font-medium shadow-sm"
          >
            Выгрузить в Excel
          </button>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <p className="text-gray-500 text-lg">Загрузка данных с сервера...</p>
          </div>
        ) : (
          <div className="bg-white rounded-[24px] border border-gray-200 shadow-sm overflow-hidden flex flex-col xl:flex-row h-[70vh]">
            
            {/* Левая часть: Таблица */}
            <div className="flex-1 overflow-y-auto xl:border-r border-gray-200 custom-scrollbar">
              <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 bg-gray-50 z-10">
                  <tr className="border-b border-gray-200">
                    <th className="py-4 px-6 text-gray-500 font-medium">Дата и время</th>
                    <th className="py-4 px-6 text-gray-500 font-medium">БИН организации</th>
                    <th className="py-4 px-6 text-gray-500 font-medium">Должность</th>
                    <th className="py-4 px-6 text-gray-500 font-medium text-center">Оценка</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {records.map((record) => {
                    const isActive = selectedRecord?.ID === record.ID;
                    return (
                      <tr 
                        key={record.ID}
                        onClick={() => setSelectedRecord(record)}
                        className={`cursor-pointer transition-colors hover:bg-gray-50 ${isActive ? 'bg-[#aff0d8]/10' : ''}`}
                      >
                        <td className="py-4 px-6 relative">
                          {isActive && <div className="absolute left-0 top-0 bottom-0 w-1 bg-[#296956] rounded-r-full"></div>}
                          <div className="font-medium text-gray-900">
                            {format(new Date(record.CreatedAt), 'dd MMM yyyy', { locale: ru })}
                          </div>
                          <div className="text-sm text-gray-500">
                            {format(new Date(record.CreatedAt), 'HH:mm')}
                          </div>
                        </td>
                        <td className="py-4 px-6 text-gray-800">{record.BIN}</td>
                        <td className="py-4 px-6 text-gray-800">{record.Position}</td>
                        <td className="py-4 px-6 text-center">
                          <span className={`inline-flex items-center justify-center w-8 h-8 rounded-full font-bold
                            ${record.Score >= 4 ? 'bg-[#aff0d8] text-[#306f5c]' : 
                              record.Score === 3 ? 'bg-yellow-200 text-yellow-800' : 'bg-[#ffdad6] text-[#93000a]'}`}>
                            {record.Score}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Правая часть: Детали выбранной анкеты */}
            <div className="w-full xl:w-[450px] bg-gray-50/50 p-6 flex flex-col">
              {selectedRecord ? (
                <>
                  <div className="mb-6 flex justify-between items-start">
                    <div>
                      <h3 className="text-xl font-bold text-[#182442] mb-1">Детали анкеты</h3>
                      <p className="text-gray-500 font-medium">БИН: {selectedRecord.BIN}</p>
                    </div>
                    <span className="inline-flex items-center justify-center px-3 py-1 rounded-full bg-[#aff0d8] text-[#306f5c] font-bold text-sm">
                      Итог: {selectedRecord.Score}
                    </span>
                  </div>
                  <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar">
                    <div className="flex flex-col gap-4">
                      {Object.entries(getParsedAnswers(selectedRecord.Answers)).map(([qKey, answer], idx) => (
                        <div key={qKey} className="bg-white p-4 rounded-xl border border-gray-100 shadow-sm">
                          <p className="text-gray-800 mb-2 font-medium text-sm">Вопрос {idx + 1}</p>
                          <div className="flex items-center gap-2">
                            <span className={`font-semibold ${String(answer).toLowerCase() === 'да' ? 'text-green-600' : String(answer).toLowerCase() === 'нет' ? 'text-red-600' : 'text-gray-500'}`}>
                              {String(answer)}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </>
              ) : (
                <div className="flex items-center justify-center h-full text-gray-400">
                  Выберите запись для просмотра деталей
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}