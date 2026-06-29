import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/useAuth'

export function Layout() {
  const { isAuthenticated, user, logout } = useAuth()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-50 text-gray-900 flex flex-col">
      <header className="border-b border-gray-200 bg-white">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
          <Link to="/" className="text-xl font-bold text-blue-600">
            auto-platform
          </Link>
          <nav className="flex items-center gap-4 text-sm">
            <Link to="/" className="hover:text-blue-600">
              Объявления
            </Link>
            {isAuthenticated ? (
              <>
                <Link to="/mine" className="hover:text-blue-600">
                  Мои объявления
                </Link>
                <Link to="/messages" className="hover:text-blue-600">
                  Сообщения
                </Link>
                <Link to="/profile" className="hover:text-blue-600">
                  {user?.username}
                </Link>
                <button onClick={handleLogout} className="text-gray-500 hover:text-red-600">
                  Выйти
                </button>
              </>
            ) : (
              <>
                <Link to="/login" className="hover:text-blue-600">
                  Войти
                </Link>
                <Link
                  to="/register"
                  className="rounded bg-blue-600 px-3 py-1.5 text-white hover:bg-blue-700"
                >
                  Регистрация
                </Link>
              </>
            )}
          </nav>
        </div>
      </header>
      <main className="mx-auto w-full max-w-5xl flex-1 px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
