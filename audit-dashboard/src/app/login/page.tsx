'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      // Отправляем запрос на Go-бэкенд
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      if (res.ok) {
        // HTTP-Only куки автоматически сохраняются в браузере
        // После успешного входа перенаправляем на дашборд
        router.push('/');
        router.refresh(); 
      } else {
        const data = await res.json().catch(() => null);
        setError(data?.message || 'Неверный пароль или логин');
      }
    } catch (err) {
      setError('Связь с сервером потеряна');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-neutral-950 flex flex-col justify-center py-12 sm:px-6 lg:px-8 font-sans">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <h2 className="mt-6 text-center text-2xl font-medium tracking-tight text-neutral-100">
          Вход в систему
        </h2>
        <p className="mt-2 text-center text-sm text-neutral-500 font-mono">
          System Authentication
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-neutral-900 py-8 px-4 border border-neutral-800 shadow sm:rounded-2xl sm:px-10">
          <form className="space-y-6" onSubmit={handleLogin}>
            {error && (
              <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20">
                <p className="text-sm text-rose-400 text-center font-mono">{error}</p>
              </div>
            )}

            <div>
              <label htmlFor="username" className="block text-sm font-medium text-neutral-300">
                Имя пользователя
              </label>
              <div className="mt-1">
                <input
                  id="username"
                  name="username"
                  type="text"
                  required
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="appearance-none block w-full px-3 py-2 border border-neutral-700 rounded-lg shadow-sm placeholder-neutral-500 bg-neutral-950 text-neutral-200 focus:outline-none focus:ring-1 focus:ring-emerald-500 focus:border-emerald-500 sm:text-sm transition-colors"
                  placeholder="admin"
                />
              </div>
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-neutral-300">
                Пароль
              </label>
              <div className="mt-1">
                <input
                  id="password"
                  name="password"
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="appearance-none block w-full px-3 py-2 border border-neutral-700 rounded-lg shadow-sm placeholder-neutral-500 bg-neutral-950 text-neutral-200 focus:outline-none focus:ring-1 focus:ring-emerald-500 focus:border-emerald-500 sm:text-sm transition-colors"
                  placeholder="••••••••"
                />
              </div>
            </div>

            <div>
              <button
                type="submit"
                disabled={isLoading}
                className="w-full flex justify-center py-2.5 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-neutral-950 bg-emerald-500 hover:bg-emerald-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-neutral-900 focus:ring-emerald-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Проверка...' : 'Вход'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}