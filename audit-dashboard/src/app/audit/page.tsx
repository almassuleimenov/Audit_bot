import { apiFetch } from '@/lib/api';
import AuditClient from './AuditClient';
//  src/app/audit/page.tsx
interface AuditRecord {
  ID: number;
  TelegramID: number;
  BIN: string;
  Position: string;
  Answers: string | Record<string, string>;
  Score: number;
  CreatedAt: string;
}

export default async function AuditPage() {
  let records: AuditRecord[] = [];
  let error: string | null = null;

  try {
    records = await apiFetch<AuditRecord[]>('/api/audits');
  } catch (err) {
    error = err instanceof Error ? err.message : 'Ошибка загрузки данных';
    console.error("Ошибка загрузки аудитов:", err);
  }

  if (error) {
    return (
      <div className="p-8 text-center text-red-500">
        <p>Ошибка загрузки данных: {error}</p>
      </div>
    );
  }

  return <AuditClient initialRecords={records} />;
}