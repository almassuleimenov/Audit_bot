// src/components/Sidebar.tsx
'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

export default function Sidebar() {
  const pathname = usePathname();

  // Простая функция для проверки активной ссылки
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

      <button className="w-full bg-[#182442] text-white font-medium py-3 rounded-xl hover:bg-[#182442]/90 transition-colors duration-200 mb-4 flex items-center justify-center gap-2 shadow-sm">
        <span className="text-xl">+</span>
        Создать отчет
      </button>

      <div className="flex flex-col gap-2 flex-grow">
        <Link 
          href="/" 
          className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium ${isActive('/') ? 'bg-[#aff0d8]/20 text-[#182442] border-l-4 border-[#aff0d8]' : 'text-gray-600 hover:bg-gray-100 hover:text-[#182442]'}`}
        >
          Обзор
        </Link>
        <Link 
          href="/audit" 
          className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium ${isActive('/audit') ? 'bg-[#aff0d8]/20 text-[#182442] border-l-4 border-[#aff0d8]' : 'text-gray-600 hover:bg-gray-100 hover:text-[#182442]'}`}
        >
          Записи аудита
        </Link>
        <Link 
          href="/appointments" 
          className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium ${isActive('/appointments') ? 'bg-[#aff0d8]/20 text-[#182442] border-l-4 border-[#aff0d8]' : 'text-gray-600 hover:bg-gray-100 hover:text-[#182442]'}`}
        >
          Заявки
        </Link>
      </div>

      <div className="mt-auto pt-4 border-t border-gray-200">
        <button className="flex items-center gap-3 px-4 py-3 rounded-xl transition-all font-medium text-gray-600 hover:bg-gray-100 w-full text-left">
          Выйти
        </button>
      </div>
    </nav>
  );
}