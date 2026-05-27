'use client';

import { useEffect, useState } from 'react';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

interface AuditRecord {
  ID: number;
  TelegramID: number;
  PhoneNumber: string;
  BIN: string;
  Position: string;
  Answers: string | Record<string, string>;
  Score: number;
  CreatedAt: string;
}

interface AuditClientProps {
  initialRecords: AuditRecord[];
}

export default function AuditClient({ initialRecords }: AuditClientProps) {
  const [records, setRecords] = useState<AuditRecord[]>(initialRecords);
  const [selectedRecord, setSelectedRecord] = useState<AuditRecord | null>(
    initialRecords.length > 0 ? initialRecords[0] : null
  );
  
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    setRecords(initialRecords);
    if (initialRecords.length > 0 && !selectedRecord) {
      setSelectedRecord(initialRecords[0]);
    }
  }, [initialRecords, selectedRecord]);

  const handleExport = () => {
    window.location.href = '/api/download';
  };

  const getParsedAnswers = (answers: string | Record<string, string>) => {
    if (typeof answers === 'string') {
      try {
        return JSON.parse(answers);
      } catch {
        return {};
      }
    }
    return answers || {};
  };

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      <header className="sticky top-0 z-50 flex justify-between items-center px-8 h-16 w-full bg-white/80 backdrop-blur-md border-b border-gray-200">
        <h2 className="font-bold text-[#182442] text-xl">Панель управления</h2>
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-full bg-gray-300 overflow-hidden border-2 border-white">
            <div className="w-full h-full bg-blue-500"></div>
          </div>
        </div>
      </header>

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

        {records.length === 0 ? (
          <div className="flex items-center justify-center h-64 bg-white rounded-[24px] border border-gray-200 shadow-sm">
            <p className="text-gray-500 text-lg">Нет данных для отображения</p>
          </div>
        ) : (
          <div className="bg-white rounded-[24px] border border-gray-200 shadow-sm overflow-hidden flex flex-col xl:flex-row h-[70vh]">
            
            {/* Левая часть: Таблица */}
            <div className="flex-1 overflow-y-auto xl:border-r border-gray-200 custom-scrollbar">
              <table className="w-full text-left border-collapse">
                <thead className="sticky top-0 bg-gray-50 z-10">
                  <tr className="border-b border-gray-200">
                    <th className="py-4 px-6 text-gray-500 font-medium">Дата и время</th>
                    <th className="py-4 px-6 text-gray-500 font-medium">Телефон</th>
                    <th className="py-4 px-6 text-gray-500 font-medium">БИН организации</th>
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
                            {isMounted ? format(new Date(record.CreatedAt), 'dd MMM yyyy', { locale: ru }) : 'Загрузка...'}
                          </div>
                          <div className="text-sm text-gray-500">
                            {isMounted ? format(new Date(record.CreatedAt), 'HH:mm') : ''}
                          </div>
                        </td>
                        <td className="py-4 px-6 text-gray-800 font-medium">{record.PhoneNumber || '—'}</td>
                        <td className="py-4 px-6 text-gray-800">{record.BIN}</td>
                        <td className="py-4 px-6 text-center">
                          <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${
                            record.Score >= 4 ? 'bg-[#aff0d8] text-[#296956]' :
                            record.Score >= 3 ? 'bg-[#bac6ec] text-[#182442]' :
                            'bg-[#ffb4a9] text-[#c41c1c]'
                          }`}>
                            {record.Score.toFixed(1)}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Правая часть: Детали */}
            {selectedRecord && (
              <div className="w-full xl:w-1/3 p-6 overflow-y-auto flex flex-col gap-6 bg-gray-50">
                <div className="flex justify-between items-center">
                    <div>
                    <h3 className="font-bold text-[#182442] mb-1">ID записи</h3>
                    <p className="text-gray-700 font-mono text-sm">#{selectedRecord.ID}</p>
                    </div>
                    <div className="inline-block px-4 py-2 rounded-full text-lg font-bold shadow-sm border"
                        style={{
                        backgroundColor: selectedRecord.Score >= 4 ? '#aff0d8' : selectedRecord.Score >= 3 ? '#bac6ec' : '#ffb4a9',
                        color: selectedRecord.Score >= 4 ? '#296956' : selectedRecord.Score >= 3 ? '#182442' : '#c41c1c',
                        borderColor: selectedRecord.Score >= 4 ? '#8ee2c3' : selectedRecord.Score >= 3 ? '#95a7d9' : '#f09689'
                        }}
                    >
                        {selectedRecord.Score.toFixed(1)} / 5.0
                    </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                    <div className="bg-white p-4 rounded-xl border border-gray-200 shadow-sm">
                        <h3 className="font-bold text-[#182442] text-xs uppercase mb-1">Телефон</h3>
                        <p className="text-gray-900 font-medium">{selectedRecord.PhoneNumber || 'Не указан'}</p>
                    </div>
                    <div className="bg-white p-4 rounded-xl border border-gray-200 shadow-sm">
                        <h3 className="font-bold text-[#182442] text-xs uppercase mb-1">БИН организации</h3>
                        <p className="text-gray-900 font-medium">{selectedRecord.BIN}</p>
                    </div>
                </div>

                <div className="bg-white p-4 rounded-xl border border-gray-200 shadow-sm">
                  <h3 className="font-bold text-[#182442] text-xs uppercase mb-1">Должность</h3>
                  <p className="text-gray-900 font-medium">{selectedRecord.Position}</p>
                </div>

                <div>
                  <h3 className="font-bold text-[#182442] mb-2">Ответы клиента</h3>
                  <div className="bg-white p-4 rounded-xl border border-gray-200 shadow-sm space-y-4">
                    {Object.entries(getParsedAnswers(selectedRecord.Answers)).map(([key, value]) => (
                      <div key={key} className="text-sm pb-3 border-b border-gray-100 last:border-0 last:pb-0">
                        <p className="text-gray-500 text-xs uppercase mb-1">{key}</p>
                        <p className="text-gray-900 font-medium">{String(value)}</p>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}