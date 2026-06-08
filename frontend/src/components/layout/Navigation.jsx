import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@stores/authStore'
import { useThemeStore } from '@stores/themeStore'
import { cn } from '@utils/cn'
import { ROUTES } from '@/routes/routes'

const navItems = [
  { href: ROUTES.DASHBOARD, label: 'Dashboard', icon: 'dashboard' },
  { href: ROUTES.DEBTS, label: 'Debts', icon: 'money' },
  { href: ROUTES.CONTACTS, label: 'Contacts', icon: 'people' },
  { href: ROUTES.NOTIFICATIONS, label: 'Notifications', icon: 'notifications' },
  { href: ROUTES.SETTINGS, label: 'Settings', icon: 'settings' },
]

const NavIcon = ({ icon }) => {
  const icons = {
    dashboard: (
      <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <rect x="3" y="3" width="7" height="7" rx="2" strokeWidth="2" stroke="currentColor" />
        <rect x="14" y="3" width="7" height="7" rx="2" strokeWidth="2" stroke="currentColor" />
        <rect x="14" y="14" width="7" height="7" rx="2" strokeWidth="2" stroke="currentColor" />
        <rect x="3" y="14" width="7" height="7" rx="2" strokeWidth="2" stroke="currentColor" />
      </svg>
    ),
    money: (
      <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"
        />
      </svg>
    ),
    people: (
      <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"
        />
      </svg>
    ),
    settings: (
      <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
        />
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
        />
      </svg>
    ),
    notifications: (
      <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
        />
      </svg>
    ),
  }

  return icons[icon] || null
}

export const Navigation = ({ mobile = false, onNavigate }) => {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((state) => state.user)
  const logout = useAuthStore((state) => state.logout)
  const isDark = useThemeStore((state) => state.isDark)
  const toggleTheme = useThemeStore((state) => state.toggle)
  console.log(user)

  const isActive = (href) => {
    // Special case for dashboard: both "/" and "/dashboard" should highlight
    if (href === ROUTES.DASHBOARD) {
      return location.pathname === '/' || location.pathname === ROUTES.DASHBOARD
    }
    return location.pathname === href
  }

  const handleNavigation = (href) => {
    if (mobile && onNavigate) {
      onNavigate()
    }
    navigate(href)
  }

  const handleLogout = () => {
    logout()
    navigate(ROUTES.LOGIN)
  }

  // Get user initials for avatar
  const userInitials =
    user && user.first_name && user.last_name ? `${user.first_name[0]}${user.last_name[0]}` : 'U'

  return (
    <nav
      className={cn(
        'z-40 h-screen w-64 border-r border-border bg-card',
        mobile ? 'fixed left-0 top-0' : 'fixed left-0 top-0'
      )}
    >
      <div className="p-6">
        {/* Logo (hide on mobile since it's in the header) */}
        {!mobile && (
          <div className="mb-8 flex items-center space-x-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
              <svg
                className="h-5 w-5 text-primary-foreground"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"
                />
              </svg>
            </div>
            <h1 className="text-xl font-bold text-card-foreground">DebtTracker</h1>
          </div>
        )}

        {/* Mobile Header Spacing */}
        {mobile && <div className="mb-8" />}

        {/* Theme Toggle */}
        <div className="mb-6">
          <button
            onClick={toggleTheme}
            className="flex w-full items-center justify-center space-x-2 rounded-lg bg-secondary px-3 py-2 text-secondary-foreground transition-colors duration-200 hover:bg-secondary/80"
          >
            {isDark ? (
              <>
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
                  />
                </svg>
                <span className="text-sm">Light Mode</span>
              </>
            ) : (
              <>
                <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
                  />
                </svg>
                <span className="text-sm">Dark Mode</span>
              </>
            )}
          </button>
        </div>

        {/* Navigation Items */}
        <ul className="space-y-2">
          {navItems.map((item) => (
            <li key={item.href}>
              <button
                onClick={() => handleNavigation(item.href)}
                className={cn(
                  'flex w-full items-center space-x-3 rounded-lg px-3 py-2 text-left text-sm font-medium transition-colors duration-200',
                  isActive(item.href)
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-secondary hover:text-secondary-foreground'
                )}
              >
                <NavIcon icon={item.icon} />
                <span>{item.label}</span>
              </button>
            </li>
          ))}
        </ul>
      </div>

      {/* User Profile */}
      <div className="absolute bottom-0 left-0 right-0 border-t border-border p-6">
        <div className="mb-3 flex items-center space-x-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary">
            <span className="text-sm font-medium text-primary-foreground">{userInitials}</span>
          </div>
          <div className="min-w-0 flex-1">
            {user && user.first_name && user.last_name ? (
              <>
                <p className="truncate text-sm font-medium text-card-foreground">
                  {user.first_name} {user.last_name}
                </p>
                <p className="truncate text-xs text-muted-foreground">{user.email}</p>
              </>
            ) : (
              <>
                <p className="truncate text-sm font-medium text-card-foreground">User</p>
                <p className="truncate text-xs text-muted-foreground">Not signed in</p>
              </>
            )}
          </div>
        </div>
        <button
          onClick={handleLogout}
          className="w-full rounded-lg px-3 py-2 text-left text-sm text-muted-foreground transition-colors duration-200 hover:bg-secondary"
        >
          Sign out
        </button>
      </div>
    </nav>
  )
}
