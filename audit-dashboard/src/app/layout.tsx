import './globals.css';
import Sidebar from '@/components/Sidebar';

export const metadata = {
  title: 'Управление - Аналитика и аудит',
  description: 'Панель управления Департамента аудита',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ru">
      <body className="bg-[#f7f9fb] text-[#191c1e] font-sans antialiased min-h-screen flex">
        {/* Боковое меню всегда на экране слева */}
        <Sidebar />
        
        {/* Основной контент страницы будет подставляться сюда */}
        <div className="flex-1 md:ml-[280px]">
          {children}
        </div>
      </body>
    </html>
  );
}