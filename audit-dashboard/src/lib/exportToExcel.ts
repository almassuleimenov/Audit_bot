import * as XLSX from 'xlsx';
// D:\Project\backend_projects\audit_bot\audit-dashboard\src\lib\exportToExcel.ts
export function exportToExcel(data: any[], filename: string) {
  if (!data || data.length === 0) {
    alert("Нет данных для экспорта");
    return;
  }

  // Форматируем данные для Excel (разворачиваем JSON с ответами)
  const formattedData = data.map((item) => {
    // Парсим ответы из JSONB
    let answers = {};
    try {
      answers = typeof item.Answers === 'string' ? JSON.parse(item.Answers) : item.Answers;
    } catch (e) {
      console.error("Ошибка парсинга ответов", e);
    }

    return {
      "ID": item.ID,
      "Telegram ID": item.TelegramID,
      "БИН": item.BIN,
      "Должность": item.Position,
      "Оценка": item.Score,
      "Дата создания": new Date(item.CreatedAt).toLocaleString('ru-RU'),
      // Динамически добавляем колонки с ответами (Q1, Q2...)
      ...answers
    };
  });

  // Создаем книгу и лист
  const worksheet = XLSX.utils.json_to_sheet(formattedData);
  const workbook = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(workbook, worksheet, "Аудит");

  // Генерируем файл и скачиваем
  XLSX.writeFile(workbook, `${filename}_${new Date().toLocaleDateString('ru-RU')}.xlsx`);
}