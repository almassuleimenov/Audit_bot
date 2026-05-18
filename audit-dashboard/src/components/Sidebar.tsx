// src/components/Sidebar.tsx
'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

export default function Sidebar() {
  const pathname = usePathname();
  const isActive = (path: string) => pathname === path;

  return (
    <nav className="hidden md:flex flex-col py-8 px-6 gap-y-4 bg-white border-r border-gray-200 shadow-none fixed left-0 top-0 h-screen w-[280px] z-40">
      <div className="flex items-center gap-4 mb-8">
        <div className="w-10 h-10 rounded-xl bg-[#dae2ff] flex items-center justify-center text-[#182442] font-bold text-xl">
          У
        </div>
        <div>
          <h1 className="font-bold text-[#182442] text-[20px] leading-tight">Управление</h1>
          <p className="text-gray-500 text-[12px]">Аналитика и аудит</p>
        </div>
      </div>

      <div className="flex flex-col gap-2 flex-grow">
        <Link href="/" className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium ${isActive('/') ? 'bg-[#aff0d8]/20 text-[#182442] border-l-4 border-[#aff0d8]' : 'text-gray-600 hover:bg-gray-100'}`}>Обзор</Link>
        <Link href="/audit" className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium ${isActive('/audit') ? 'bg-[#aff0d8]/20 text-[#182442] border-l-4 border-[#aff0d8]' : 'text-gray-600 hover:bg-gray-100'}`}>Записи аудита</Link>
        <Link href="/appointments" className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium ${isActive('/appointments') ? 'bg-[#aff0d8]/20 text-[#182442] border-l-4 border-[#aff0d8]' : 'text-gray-600 hover:bg-gray-100'}`}>Заявки</Link>
        
        {/* НОВАЯ КНОПКА */}
        <Link href="/settings" className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium ${isActive('/settings') ? 'bg-[#aff0d8]/20 text-[#182442] border-l-4 border-[#aff0d8]' : 'text-gray-600 hover:bg-gray-100'}`}>Настройки вопросов</Link>
      </div>
    </nav>
  );
}