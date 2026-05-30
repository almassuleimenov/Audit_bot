export const dynamic = "force-dynamic";
import { apiFetch } from '@/lib/api';

interface Appointment {
  ID: number;
  FullName: string;
  PhoneNumber: string;
  TargetManager: string;
  Question: string;
  CreatedAt: string;
}

export default async function AppointmentsPage() {
  let appointments: Appointment[] = [];
  let error: string | null = null;

  try {
    const data = await apiFetch<Appointment[]>('/api/appointments');
    
    // Защита от nil-слайса со стороны Go (когда БД пустая, возвращается null)
    if (data && Array.isArray(data)) {
      appointments = data;
    }
  } catch (err) {
    error = err instanceof Error ? err.message : 'Ошибка загрузки данных';
    console.error("Ошибка загрузки заявок:", err);
  }

  if (error) {
    return (
      <div className="p-8 text-center text-red-500">
        <p>Ошибка загрузки данных: {error}</p>
      </div>
    );
  }

  return (
    <main className="flex flex-col min-h-screen">
      <header className="sticky top-0 z-50 flex justify-between items-center px-8 h-16 bg-white/70 backdrop-blur-md border-b border-gray-200">
        <h2 className="font-bold text-[#182442] text-xl">Панель управления</h2>
      </header>

      <div className="p-8 max-w-7xl mx-auto w-full flex flex-col gap-8">
        <div>
          <h2 className="text-3xl font-bold text-[#182442] mb-2">Онлайн-приемная</h2>
          <p className="text-gray-500">Управление заявками на встречи и консультации</p>
        </div>

        {appointments.length === 0 ? (
          <p className="text-gray-500">Нет заявок на встречи</p>
        ) : (
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
            {appointments.map((appt) => (
              <div key={appt.ID} className="bg-white rounded-[24px] p-6 shadow-sm border border-gray-100 flex flex-col gap-4 border-l-4 border-l-[#ffb4a9]">
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="text-xl font-bold text-[#182442] mb-1">{appt.FullName}</h3>
                    <p className="text-gray-500 text-sm">
                      {new Date(appt.CreatedAt).toLocaleString('ru-RU')}
                    </p>
                  </div>
                </div>

                <div className="bg-gray-50 p-4 rounded-xl border border-gray-200">
                  <span className="text-gray-500 text-sm block mb-1">К кому записан</span>
                  <span className="font-medium text-[#182442]">{appt.TargetManager}</span>
                </div>

                <div>
                  <span className="text-gray-500 text-sm block mb-1">Характер вопроса</span>
                  <p className="text-[#191c1e] bg-white rounded-lg">{appt.Question}</p>
                </div>

                <div className="mt-auto pt-4 border-t border-gray-100">
                  <a 
                    href={`tel:${appt.PhoneNumber}`} 
                    className="inline-flex items-center gap-2 text-[#296956] font-semibold hover:underline"
                  >
                    📞 Позвонить: {appt.PhoneNumber}
                  </a>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </main>
  );
}