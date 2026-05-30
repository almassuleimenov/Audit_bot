export const dynamic = "force-dynamic";
import { apiFetch } from '@/lib/api';
import AuditClient from './AuditClient';

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

export default async function AuditPage() {
  let records: AuditRecord[] = [];
  
  try {
    const data = await apiFetch<AuditRecord[]>('/api/audits', {
      next: { revalidate: 30 }
    });
    
    if (data && Array.isArray(data)) {
      records = data;
    }
  } catch (error) {
    console.error('Ошибка загрузки данных аудита:', error);
  }

  return <AuditClient initialRecords={records} />;
}